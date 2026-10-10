package com.fujiwaracreative.famloop.ui.parent

import android.annotation.SuppressLint
import android.os.Bundle
import android.widget.Toast
import androidx.appcompat.app.AppCompatActivity
import androidx.camera.core.*
import androidx.camera.lifecycle.ProcessCameraProvider
import androidx.camera.view.PreviewView
import androidx.core.content.ContextCompat
import com.fujiwaracreative.famloop.data.remote.NetworkClient
import com.fujiwaracreative.famloop.data.remote.PairDeviceRequest
import com.google.mlkit.vision.barcode.BarcodeScanning
import com.google.mlkit.vision.common.InputImage
import kotlinx.coroutines.CoroutineScope
import kotlinx.coroutines.Dispatchers
import kotlinx.coroutines.launch
import org.json.JSONObject
import java.util.concurrent.Executors

class QrScannerActivity : AppCompatActivity() {

    private lateinit var previewView: PreviewView
    private val cameraExecutor = Executors.newSingleThreadExecutor()
    private var isScanned = false
    private var targetGroupId: String = ""

    override fun onCreate(savedInstanceState: Bundle?) {
        super.onCreate(savedInstanceState)
        targetGroupId = intent.getStringExtra("EXTRA_GROUP_ID") ?: ""

        previewView = PreviewView(this)
        setContentView(previewView)

        startCamera()
    }

    private fun startCamera() {
        val cameraProviderFuture = ProcessCameraProvider.getInstance(this)
        cameraProviderFuture.addListener({
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
            try {
                cameraProvider.unbindAll()
                cameraProvider.bindToLifecycle(this, cameraSelector, preview, imageAnalyzer)
            } catch (e: Exception) {
                Toast.makeText(this, "Gagal membuka kamera: ${e.message}", Toast.LENGTH_SHORT).show()
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

            CoroutineScope(Dispatchers.Main).launch {
                Toast.makeText(this@QrScannerActivity, "Menghubungkan $deviceName...", Toast.LENGTH_SHORT).show()
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
                    if (resp.isSuccessful) {
                        Toast.makeText(this@QrScannerActivity, "Berhasil menghubungkan $deviceName!", Toast.LENGTH_LONG).show()
                        finish()
                    } else {
                        Toast.makeText(this@QrScannerActivity, "Gagal pairing: ${resp.errorBody()?.string()}", Toast.LENGTH_LONG).show()
                        isScanned = false
                    }
                } catch (e: Exception) {
                    Toast.makeText(this@QrScannerActivity, "Error koneksi server: ${e.message}", Toast.LENGTH_LONG).show()
                    isScanned = false
                }
            }
        } catch (e: Exception) {
            runOnUiThread {
                Toast.makeText(this, "Format QR tidak valid: bukan perangkat FamLoop", Toast.LENGTH_SHORT).show()
                isScanned = false
            }
        }
    }

    override fun onDestroy() {
        super.onDestroy()
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
