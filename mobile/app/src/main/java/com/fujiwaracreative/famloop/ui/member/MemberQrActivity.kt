package com.fujiwaracreative.famloop.ui.member

import android.content.Intent
import android.graphics.Bitmap
import android.graphics.Color
import android.graphics.Typeface
import android.graphics.drawable.GradientDrawable
import android.os.Bundle
import android.view.Gravity
import android.view.View
import android.widget.Button
import android.widget.FrameLayout
import android.widget.ImageView
import android.widget.LinearLayout
import android.widget.ScrollView
import android.widget.TextView
import android.widget.Toast
import androidx.appcompat.app.AppCompatActivity
import com.fujiwaracreative.famloop.data.local.SessionManager
import com.fujiwaracreative.famloop.ui.MainActivity
import com.fujiwaracreative.famloop.utils.DeviceUtils
import com.fujiwaracreative.famloop.utils.NetworkUtils
import com.google.zxing.BarcodeFormat
import com.journeyapps.barcodescanner.BarcodeEncoder
import org.json.JSONObject

class MemberQrActivity : AppCompatActivity() {

    override fun onCreate(savedInstanceState: Bundle?) {
        super.onCreate(savedInstanceState)
        SessionManager.getInstance(this).role = "member"

        val scrollView = ScrollView(this).apply {
            setBackgroundColor(Color.parseColor("#F8FAFC"))
            isFillViewport = true
        }

        val root = LinearLayout(this).apply {
            orientation = LinearLayout.VERTICAL
            gravity = Gravity.CENTER_HORIZONTAL
            setPadding(40, 56, 40, 56)
        }

        // Top Navigation & Context
        val topBar = LinearLayout(this).apply {
            orientation = LinearLayout.HORIZONTAL
            gravity = Gravity.CENTER_VERTICAL
            setPadding(0, 0, 0, 32)
            layoutParams = LinearLayout.LayoutParams(
                LinearLayout.LayoutParams.MATCH_PARENT,
                LinearLayout.LayoutParams.WRAP_CONTENT
            )
        }

        val btnBack = TextView(this).apply {
            text = "←"
            textSize = 22f
            setTypeface(null, Typeface.BOLD)
            setTextColor(Color.parseColor("#0F172A"))
            setPadding(16, 12, 16, 12)
            setOnClickListener { finish() }
        }
        topBar.addView(btnBack)

        val badgeContainer = LinearLayout(this).apply {
            gravity = Gravity.CENTER
            layoutParams = LinearLayout.LayoutParams(0, LinearLayout.LayoutParams.WRAP_CONTENT, 1f)
        }
        val setupBadge = TextView(this).apply {
            text = "🧒 CHILD PROFILE SETUP"
            textSize = 11f
            setTypeface(null, Typeface.BOLD)
            setTextColor(Color.parseColor("#004AC6"))
            val bg = GradientDrawable().apply {
                setColor(Color.parseColor("#E0EAFF"))
                cornerRadius = 9999f
            }
            background = bg
            setPadding(28, 10, 28, 10)
        }
        badgeContainer.addView(setupBadge)
        topBar.addView(badgeContainer)

        val spacerRight = TextView(this).apply {
            text = "   "
            textSize = 18f
        }
        topBar.addView(spacerRight)
        root.addView(topBar)

        // Hero Profile Anchor
        val heroLayout = LinearLayout(this).apply {
            orientation = LinearLayout.VERTICAL
            gravity = Gravity.CENTER_HORIZONTAL
            setPadding(0, 0, 0, 36)
        }

        val avatarBox = TextView(this).apply {
            text = "🧒"
            textSize = 36f
            gravity = Gravity.CENTER
            val bg = GradientDrawable().apply {
                setColor(Color.parseColor("#EFF6FF"))
                setStroke(4, Color.parseColor("#3B82F6"))
                cornerRadius = 9999f
            }
            background = bg
            layoutParams = LinearLayout.LayoutParams(160, 160)
        }
        heroLayout.addView(avatarBox)

        val deviceModel = DeviceUtils.getDeviceModelName()
        val titleView = TextView(this).apply {
            text = "Hubungkan $deviceModel"
            textSize = 22f
            setTypeface(null, Typeface.BOLD)
            gravity = Gravity.CENTER
            setTextColor(Color.parseColor("#0F172A"))
            setPadding(0, 20, 0, 8)
        }
        heroLayout.addView(titleView)

        val subtitleView = TextView(this).apply {
            text = "Minta Orang Tua memindai kode QR ini menggunakan kamera FamLoop di HP mereka."
            textSize = 13f
            gravity = Gravity.CENTER
            setTextColor(Color.parseColor("#64748B"))
            setLineSpacing(3f, 1f)
        }
        heroLayout.addView(subtitleView)
        root.addView(heroLayout)

        // Minimalist Member Pairing Card with Google Stitch Brackets
        val cardLayout = LinearLayout(this).apply {
            orientation = LinearLayout.VERTICAL
            gravity = Gravity.CENTER_HORIZONTAL
            setPadding(36, 40, 36, 40)
            val bg = GradientDrawable().apply {
                setColor(Color.WHITE)
                cornerRadius = 32f
                setStroke(2, Color.parseColor("#E2E8F0"))
            }
            background = bg
            layoutParams = LinearLayout.LayoutParams(
                LinearLayout.LayoutParams.MATCH_PARENT,
                LinearLayout.LayoutParams.WRAP_CONTENT
            )
        }

        // Ready to Sync Status Pill
        val statusPill = TextView(this).apply {
            text = "●  READY TO SYNC"
            textSize = 11f
            setTypeface(null, Typeface.BOLD)
            setTextColor(Color.parseColor("#059669"))
            val bg = GradientDrawable().apply {
                setColor(Color.parseColor("#ECFDF5"))
                cornerRadius = 9999f
            }
            background = bg
            setPadding(28, 10, 28, 10)
        }
        cardLayout.addView(statusPill)

        // QR Code Container Frame with Royal Blue precision corner brackets
        val qrContainer = FrameLayout(this).apply {
            val bg = GradientDrawable().apply {
                setColor(Color.WHITE)
                cornerRadius = 24f
                setStroke(4, Color.parseColor("#2563EB")) // Precision boundary
            }
            background = bg
            setPadding(24, 24, 24, 24)
            val size = (resources.displayMetrics.widthPixels * 0.65).toInt()
            layoutParams = LinearLayout.LayoutParams(size, size).apply {
                topMargin = 32
                bottomMargin = 24
            }
        }

        val qrImageView = ImageView(this).apply {
            layoutParams = FrameLayout.LayoutParams(
                FrameLayout.LayoutParams.MATCH_PARENT,
                FrameLayout.LayoutParams.MATCH_PARENT
            )
        }
        qrContainer.addView(qrImageView)
        cardLayout.addView(qrContainer)

        val infoView = TextView(this).apply {
            textSize = 12f
            gravity = Gravity.CENTER
            setTextColor(Color.parseColor("#94A3B8"))
            setPadding(0, 12, 0, 8)
        }
        cardLayout.addView(infoView)
        root.addView(cardLayout)

        // Action: Buka Peta Bersama
        val mapBtn = Button(this).apply {
            text = "Buka Peta Bersama"
            setTextColor(Color.WHITE)
            textSize = 14f
            setTypeface(null, Typeface.BOLD)
            val bg = GradientDrawable().apply {
                setColor(Color.parseColor("#059669"))
                cornerRadius = 20f
            }
            background = bg
            setPadding(32, 24, 32, 24)
            layoutParams = LinearLayout.LayoutParams(
                LinearLayout.LayoutParams.MATCH_PARENT,
                LinearLayout.LayoutParams.WRAP_CONTENT
            ).apply {
                topMargin = 36
            }
            setOnClickListener {
                startActivity(Intent(this@MemberQrActivity, MemberMapActivity::class.java))
            }
        }
        root.addView(mapBtn)

        // Action: Ganti Peran
        val switchBtn = Button(this).apply {
            text = "← Ganti Peran Perangkat"
            setTextColor(Color.parseColor("#64748B"))
            setBackgroundColor(Color.TRANSPARENT)
            textSize = 13f
            setPadding(0, 24, 0, 16)
            setOnClickListener {
                val intent = Intent(this@MemberQrActivity, MainActivity::class.java).apply {
                    putExtra("EXTRA_FORCE_CHOOSER", true)
                }
                startActivity(intent)
                finish()
            }
        }
        root.addView(switchBtn)

        scrollView.addView(root)
        setContentView(scrollView)

        generateAndDisplayQR(qrImageView, infoView)
    }

    private fun generateAndDisplayQR(imageView: ImageView, infoView: TextView) {
        try {
            val deviceId = DeviceUtils.getHardwareDeviceId(this)
            val modelName = DeviceUtils.getDeviceModelName()
            val ips = NetworkUtils.getDeviceIPAddresses()

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
            infoView.text = "Model: $modelName\nHardware UUID: ${deviceId.take(12)}...\nIP: ${ips.ipv4 ?: "Offline"}"
        } catch (e: Exception) {
            Toast.makeText(this, "Gagal membuat QR: ${e.message}", Toast.LENGTH_LONG).show()
        }
    }
}
