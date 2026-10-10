package com.fujiwaracreative.famloop.ui.member

import android.graphics.Bitmap
import android.os.Bundle
import android.widget.ImageView
import android.widget.TextView
import android.widget.Toast
import androidx.appcompat.app.AppCompatActivity
import com.fujiwaracreative.famloop.utils.DeviceUtils
import com.fujiwaracreative.famloop.utils.NetworkUtils
import com.google.zxing.BarcodeFormat
import com.journeyapps.barcodescanner.BarcodeEncoder
import org.json.JSONObject

class MemberQrActivity : AppCompatActivity() {

    override fun onCreate(savedInstanceState: Bundle?) {
        super.onCreate(savedInstanceState)

        // Simple programmatic view layout
        val root = android.widget.LinearLayout(this).apply {
            orientation = android.widget.LinearLayout.VERTICAL
            gravity = android.view.Gravity.CENTER
            setPadding(48, 48, 48, 48)
            setBackgroundColor(android.graphics.Color.WHITE)
        }

        val titleView = TextView(this).apply {
            text = "QR Hubungkan Perangkat Keluarga"
            textSize = 20f
            setTypeface(null, android.graphics.Typeface.BOLD)
            gravity = android.view.Gravity.CENTER
            setTextColor(android.graphics.Color.parseColor("#1E293B"))
        }

        val subtitleView = TextView(this).apply {
            text = "Arahkan kamera HP Orang Tua / Kepala Keluarga ke QR code di bawah ini"
            textSize = 14f
            gravity = android.view.Gravity.CENTER
            setTextColor(android.graphics.Color.parseColor("#64748B"))
            setPadding(0, 16, 0, 32)
        }

        val qrImageView = ImageView(this).apply {
            layoutParams = android.widget.LinearLayout.LayoutParams(600, 600)
        }

        val infoView = TextView(this).apply {
            textSize = 12f
            gravity = android.view.Gravity.CENTER
            setTextColor(android.graphics.Color.parseColor("#94A3B8"))
            setPadding(0, 32, 0, 0)
        }

        root.addView(titleView)
        root.addView(subtitleView)
        root.addView(qrImageView)
        root.addView(infoView)
        setContentView(root)

        generateAndDisplayQR(qrImageView, infoView)
    }

    private fun generateAndDisplayQR(imageView: ImageView, infoView: TextView) {
        try {
            val deviceId = DeviceUtils.getHardwareDeviceId(this)
            val modelName = DeviceUtils.getDeviceModelName()
            val ips = NetworkUtils.getDeviceIPAddresses()

            // Payload JSON yang diikat langsung ke hardware perangkat
            val qrJson = JSONObject().apply {
                put("protocol", "famloop_pair_v1")
                put("device_uuid", deviceId)
                put("device_name", modelName)
                put("ipv4", ips.ipv4 ?: "")
                put("ipv6", ips.ipv6 ?: "")
                put("timestamp", System.currentTimeMillis())
            }

            val encoder = BarcodeEncoder()
            val bitmap: Bitmap = encoder.encodeBitmap(
                qrJson.toString(),
                BarcodeFormat.QR_CODE,
                600,
                600
            )

            imageView.setImageBitmap(bitmap)
            infoView.text = "Perangkat: $modelName\nID: ${deviceId.take(8)}... | IP: ${ips.ipv4 ?: "Offline"}"
        } catch (e: Exception) {
            Toast.makeText(this, "Gagal membuat QR: ${e.message}", Toast.LENGTH_LONG).show()
        }
    }
}
