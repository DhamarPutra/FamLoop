package com.fujiwaracreative.famloop.ui.parent

import android.Manifest
import android.annotation.SuppressLint
import android.content.pm.PackageManager
import android.graphics.Color
import android.graphics.Typeface
import android.graphics.drawable.GradientDrawable
import android.os.Bundle
import android.view.Gravity
import android.view.View
import android.widget.Button
import android.widget.FrameLayout
import android.widget.LinearLayout
import android.widget.TextView
import android.widget.Toast
import androidx.activity.result.contract.ActivityResultContracts
import androidx.appcompat.app.AppCompatActivity
import androidx.camera.core.*
import androidx.camera.lifecycle.ProcessCameraProvider
import androidx.camera.view.PreviewView
import androidx.core.content.ContextCompat
import com.fujiwaracreative.famloop.data.local.SessionManager
import com.fujiwaracreative.famloop.data.remote.NetworkClient
import com.fujiwaracreative.famloop.data.remote.PairDeviceRequest
import com.google.mlkit.vision.barcode.BarcodeScanning
import com.google.mlkit.vision.common.InputImage
import kotlinx.coroutines.CoroutineScope
import kotlinx.coroutines.Dispatchers
import kotlinx.coroutines.launch
import kotlinx.coroutines.withContext
import org.json.JSONObject
import java.util.concurrent.Executors

class QrScannerActivity : AppCompatActivity() {

    private lateinit var previewView: PreviewView
    private lateinit var permissionLayout: LinearLayout
    private val cameraExecutor = Executors.newSingleThreadExecutor()
    private var camera: Camera? = null
    private var isTorchOn = false
    private var isScanned = false
    private var targetGroupId: String = ""
    private lateinit var session: SessionManager

    private val cameraPermissionLauncher =
        registerForActivityResult(ActivityResultContracts.RequestPermission()) { isGranted ->
            if (isGranted) {
                permissionLayout.visibility = View.GONE
                previewView.visibility = View.VISIBLE
                startCamera()
            } else {
                permissionLayout.visibility = View.VISIBLE
                previewView.visibility = View.GONE
                Toast.makeText(this, "Izin kamera diperlukan untuk memindai QR code", Toast.LENGTH_LONG).show()
            }
        }

    override fun onCreate(savedInstanceState: Bundle?) {
        super.onCreate(savedInstanceState)
        session = SessionManager.getInstance(this)
        session.token?.let { NetworkClient.authToken = it }

        targetGroupId = intent.getStringExtra("EXTRA_GROUP_ID") ?: (session.groupId ?: "")

        val rootLayout = FrameLayout(this).apply {
            setBackgroundColor(Color.parseColor("#0B1C30"))
        }

        // 1. Camera Preview View (Fill)
        previewView = PreviewView(this).apply {
            layoutParams = FrameLayout.LayoutParams(
                FrameLayout.LayoutParams.MATCH_PARENT,
                FrameLayout.LayoutParams.MATCH_PARENT
            )
        }
        rootLayout.addView(previewView)

        // 2. Vignette / Dark HUD Gradient Overlay
        val hudOverlay = View(this).apply {
            val bg = GradientDrawable().apply {
                setColor(Color.parseColor("#66000000")) // Semi-transparent
            }
            background = bg
            layoutParams = FrameLayout.LayoutParams(
                FrameLayout.LayoutParams.MATCH_PARENT,
                FrameLayout.LayoutParams.MATCH_PARENT
            )
        }
        rootLayout.addView(hudOverlay)

        // 3. Telemetry Watermark (Top Left)
        val watermark = TextView(this).apply {
            text = "● LIVE CAM • 60 FPS • AI-OPTIC v4.2"
            textSize = 10f
            setTypeface(null, Typeface.BOLD)
            setTextColor(Color.parseColor("#4CD7F6"))
            setPadding(36, 40, 0, 0)
        }
        rootLayout.addView(watermark)

        // 4. Top Navigation Bar (Close, Title, Torch)
        val topBar = LinearLayout(this).apply {
            orientation = LinearLayout.HORIZONTAL
            gravity = Gravity.CENTER_VERTICAL
            setPadding(32, 80, 32, 24)
            layoutParams = FrameLayout.LayoutParams(
                FrameLayout.LayoutParams.MATCH_PARENT,
                FrameLayout.LayoutParams.WRAP_CONTENT
            ).apply {
                gravity = Gravity.TOP
            }
        }

        val btnClose = Button(this).apply {
            text = "✕"
            setTextColor(Color.WHITE)
            textSize = 18f
            setTypeface(null, Typeface.BOLD)
            val bg = GradientDrawable().apply {
                setColor(Color.parseColor("#44000000"))
                cornerRadius = 9999f
            }
            background = bg
            layoutParams = LinearLayout.LayoutParams(96, 96)
            setOnClickListener { finish() }
        }
        topBar.addView(btnClose)

        val titleContainer = LinearLayout(this).apply {
            orientation = LinearLayout.VERTICAL
            gravity = Gravity.CENTER_HORIZONTAL
            layoutParams = LinearLayout.LayoutParams(0, LinearLayout.LayoutParams.WRAP_CONTENT, 1f)
        }
        val tTitle = TextView(this).apply {
            text = "Scan Member Device"
            textSize = 16f
            setTypeface(null, Typeface.BOLD)
            setTextColor(Color.WHITE)
        }
        val tSub = TextView(this).apply {
            text = "Radar Beacon Active"
            textSize = 11f
            setTextColor(Color.parseColor("#4CD7F6"))
        }
        titleContainer.addView(tTitle)
        titleContainer.addView(tSub)
        topBar.addView(titleContainer)

        val btnTorch = Button(this).apply {
            text = "⚡"
            setTextColor(Color.WHITE)
            textSize = 16f
            val bg = GradientDrawable().apply {
                setColor(Color.parseColor("#44000000"))
                cornerRadius = 9999f
            }
            background = bg
            layoutParams = LinearLayout.LayoutParams(96, 96)
            setOnClickListener {
                toggleTorch()
            }
        }
        topBar.addView(btnTorch)
        rootLayout.addView(topBar)

        // 5. Instruction Pill Badge
        val pillContainer = LinearLayout(this).apply {
            gravity = Gravity.CENTER
            layoutParams = FrameLayout.LayoutParams(
                FrameLayout.LayoutParams.MATCH_PARENT,
                FrameLayout.LayoutParams.WRAP_CONTENT
            ).apply {
                topMargin = 220
            }
        }
        val instructPill = TextView(this).apply {
            text = "Arahkan kamera ke kode QR di HP Anak"
            textSize = 12f
            setTypeface(null, Typeface.BOLD)
            setTextColor(Color.WHITE)
            val bg = GradientDrawable().apply {
                setColor(Color.parseColor("#880F172A"))
                cornerRadius = 9999f
                setStroke(2, Color.parseColor("#38BDF8"))
            }
            background = bg
            setPadding(36, 14, 36, 14)
        }
        pillContainer.addView(instructPill)
        rootLayout.addView(pillContainer)

        // 6. Viewfinder Reticle Aperture (Corner brackets styling)
        val reticleContainer = FrameLayout(this).apply {
            val size = (resources.displayMetrics.widthPixels * 0.72).toInt()
            layoutParams = FrameLayout.LayoutParams(size, size).apply {
                gravity = Gravity.CENTER
            }
            val bg = GradientDrawable().apply {
                setColor(Color.TRANSPARENT)
                setStroke(4, Color.parseColor("#57DFFE")) // Electric Cyan
                cornerRadius = 36f
            }
            background = bg
        }

        // Crosshair center dot
        val crosshair = View(this).apply {
            val bg = GradientDrawable().apply {
                setColor(Color.parseColor("#57DFFE"))
                cornerRadius = 9999f
            }
            background = bg
            layoutParams = FrameLayout.LayoutParams(16, 16).apply {
                gravity = Gravity.CENTER
            }
        }
        reticleContainer.addView(crosshair)
        rootLayout.addView(reticleContainer)

        // 7. Permission Denied Fallback UI
        permissionLayout = LinearLayout(this).apply {
            orientation = LinearLayout.VERTICAL
            gravity = Gravity.CENTER
            setPadding(48, 48, 48, 48)
            setBackgroundColor(Color.parseColor("#0F172A"))
            visibility = View.GONE
            layoutParams = FrameLayout.LayoutParams(
                FrameLayout.LayoutParams.MATCH_PARENT,
                FrameLayout.LayoutParams.MATCH_PARENT
            )

            val pText = TextView(this@QrScannerActivity).apply {
                text = "Akses Kamera Diperlukan\nFamLoop membutuhkan izin kamera untuk memindai QR code perangkat."
                textSize = 14f
                setTextColor(Color.WHITE)
                gravity = Gravity.CENTER
                setPadding(0, 0, 0, 36)
            }
            addView(pText)

            val pBtn = Button(this@QrScannerActivity).apply {
                text = "Izinkan Kamera"
                setTextColor(Color.WHITE)
                val bg = GradientDrawable().apply {
                    setColor(Color.parseColor("#2563EB"))
                    cornerRadius = 16f
                }
                background = bg
                setPadding(32, 20, 32, 20)
                setOnClickListener { checkAndRequestCameraPermission() }
            }
            addView(pBtn)
        }
        rootLayout.addView(permissionLayout)

        setContentView(rootLayout)

        checkAndRequestCameraPermission()
    }

    private fun toggleTorch() {
        if (camera != null && camera!!.cameraInfo.hasFlashUnit()) {
            isTorchOn = !isTorchOn
            camera!!.cameraControl.enableTorch(isTorchOn)
        } else {
            Toast.makeText(this, "Flashlight tidak tersedia", Toast.LENGTH_SHORT).show()
        }
    }

    private fun checkAndRequestCameraPermission() {
        if (ContextCompat.checkSelfPermission(this, Manifest.permission.CAMERA) == PackageManager.PERMISSION_GRANTED) {
            permissionLayout.visibility = View.GONE
            previewView.visibility = View.VISIBLE
            startCamera()
        } else {
            cameraPermissionLauncher.launch(Manifest.permission.CAMERA)
        }
    }

    private fun startCamera() {
        val cameraProviderFuture = ProcessCameraProvider.getInstance(this)
        cameraProviderFuture.addListener({
            try {
                val cameraProvider = cameraProviderFuture.get()

                val preview = Preview.Builder().build().also {
                    it.setSurfaceProvider(previewView.surfaceProvider)
                }

                val imageAnalyzer = ImageAnalysis.Builder()
                    .setBackpressureStrategy(ImageAnalysis.STRATEGY_KEEP_ONLY_LATEST)
                    .build()
                    .also {
                        it.setAnalyzer(cameraExecutor, BarcodeAnalyzer { qrRawText ->
                            if (!isScanned) {
                                isScanned = true
                                processScannedQr(qrRawText)
                            }
                        })
                    }

                val cameraSelector = CameraSelector.DEFAULT_BACK_CAMERA
                cameraProvider.unbindAll()
                camera = cameraProvider.bindToLifecycle(this, cameraSelector, preview, imageAnalyzer)
            } catch (e: Exception) {
                runOnUiThread {
                    Toast.makeText(this, "Gagal memulai kamera: ${e.message}", Toast.LENGTH_SHORT).show()
                }
            }
        }, ContextCompat.getMainExecutor(this))
    }

    private fun processScannedQr(qrText: String) {
        try {
            val json = JSONObject(qrText)
            val deviceUUID = json.getString("device_uuid")
            val deviceName = json.getString("device_name")
            val ipv4 = json.optString("ipv4", "")
            val ipv6 = json.optString("ipv6", "")

            if (targetGroupId.isEmpty()) {
                runOnUiThread {
                    Toast.makeText(this, "Group ID tidak ditemukan. Silakan login kembali.", Toast.LENGTH_LONG).show()
                    finish()
                }
                return
            }

            CoroutineScope(Dispatchers.IO).launch {
                withContext(Dispatchers.Main) {
                    Toast.makeText(this@QrScannerActivity, "Menghubungkan $deviceName ke grup...", Toast.LENGTH_SHORT).show()
                }
                try {
                    val resp = NetworkClient.apiService.pairDevice(
                        groupId = targetGroupId,
                        request = PairDeviceRequest(
                            device_uuid = deviceUUID,
                            device_name = deviceName,
                            ipv4 = ipv4,
                            ipv6 = ipv6
                        )
                    )
                    withContext(Dispatchers.Main) {
                        if (resp.isSuccessful) {
                            Toast.makeText(this@QrScannerActivity, "Berhasil menghubungkan $deviceName!", Toast.LENGTH_LONG).show()
                            setResult(RESULT_OK)
                            finish()
                        } else {
                            val err = resp.errorBody()?.string() ?: "Gagal pairing"
                            Toast.makeText(this@QrScannerActivity, "Gagal: $err", Toast.LENGTH_LONG).show()
                            isScanned = false
                        }
                    }
                } catch (e: Exception) {
                    withContext(Dispatchers.Main) {
                        Toast.makeText(this@QrScannerActivity, "Error koneksi server: ${e.message}", Toast.LENGTH_LONG).show()
                        isScanned = false
                    }
                }
            }
        } catch (_: Exception) {
            runOnUiThread {
                Toast.makeText(this, "Format QR tidak valid: bukan perangkat FamLoop", Toast.LENGTH_SHORT).show()
                isScanned = false
            }
        }
    }

    override fun onDestroy() {
        super.onDestroy()
        if (isTorchOn) {
            camera?.cameraControl?.enableTorch(false)
        }
        cameraExecutor.shutdown()
    }

    private class BarcodeAnalyzer(private val onBarcodeFound: (String) -> Unit) : ImageAnalysis.Analyzer {
        private val scanner = BarcodeScanning.getClient()

        @SuppressLint("UnsafeOptInUsageError")
        override fun analyze(imageProxy: ImageProxy) {
            val mediaImage = imageProxy.image
            if (mediaImage != null) {
                val image = InputImage.fromMediaImage(mediaImage, imageProxy.imageInfo.rotationDegrees)
                scanner.process(image)
                    .addOnSuccessListener { barcodes ->
                        for (barcode in barcodes) {
                            val raw = barcode.rawValue
                            if (raw != null) {
                                onBarcodeFound(raw)
                                break
                            }
                        }
                    }
                    .addOnCompleteListener {
                        imageProxy.close()
                    }
            } else {
                imageProxy.close()
            }
        }
    }
}
