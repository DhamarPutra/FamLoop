package com.fujiwaracreative.famloop.ui

import android.content.Intent
import android.graphics.Color
import android.graphics.Typeface
import android.graphics.drawable.GradientDrawable
import android.os.Bundle
import android.view.Gravity
import android.widget.Button
import android.widget.LinearLayout
import android.widget.ScrollView
import android.widget.TextView
import androidx.appcompat.app.AppCompatActivity
import com.fujiwaracreative.famloop.data.local.SessionManager
import com.fujiwaracreative.famloop.data.remote.NetworkClient
import com.fujiwaracreative.famloop.ui.member.MemberQrActivity
import com.fujiwaracreative.famloop.ui.parent.LiveMapActivity
import com.fujiwaracreative.famloop.ui.parent.ParentAuthActivity

class MainActivity : AppCompatActivity() {

    private lateinit var session: SessionManager
    private var selectedRole: String = "parent" // "parent" or "member"

    // UI elements to update on toggle
    private lateinit var parentCard: LinearLayout
    private lateinit var memberCard: LinearLayout
    private lateinit var continueBtn: Button

    override fun onCreate(savedInstanceState: Bundle?) {
        super.onCreate(savedInstanceState)
        session = SessionManager.getInstance(this)

        session.token?.let {
            NetworkClient.authToken = it
        }

        val forceChooser = intent.getBooleanExtra("EXTRA_FORCE_CHOOSER", false)
        if (!forceChooser) {
            if (session.role == "parent" && session.isLoggedIn && !session.groupId.isNullOrEmpty()) {
                val intent = Intent(this, LiveMapActivity::class.java).apply {
                    putExtra("EXTRA_GROUP_ID", session.groupId)
                    putExtra("EXTRA_GROUP_NAME", session.groupName)
                }
                startActivity(intent)
                finish()
                return
            } else if (session.role == "member") {
                startActivity(Intent(this, MemberQrActivity::class.java))
                finish()
                return
            }
        }

        buildStitchRoleChooserView()
    }

    private fun buildStitchRoleChooserView() {
        val scrollView = ScrollView(this).apply {
            setBackgroundColor(Color.parseColor("#F8FAFC"))
            isFillViewport = true
        }

        val root = LinearLayout(this).apply {
            orientation = LinearLayout.VERTICAL
            setPadding(48, 64, 48, 64)
        }

        // Header: Logo & Identity
        val headerLayout = LinearLayout(this).apply {
            orientation = LinearLayout.HORIZONTAL
            gravity = Gravity.CENTER_VERTICAL
            setPadding(0, 0, 0, 36)
        }

        val logoBox = TextView(this).apply {
            text = "📍"
            textSize = 24f
            gravity = Gravity.CENTER
            val bg = GradientDrawable().apply {
                setColor(Color.parseColor("#2563EB"))
                cornerRadius = 24f
            }
            background = bg
            layoutParams = LinearLayout.LayoutParams(96, 96)
        }
        headerLayout.addView(logoBox)

        val headerTextLayout = LinearLayout(this).apply {
            orientation = LinearLayout.VERTICAL
            setPadding(24, 0, 0, 0)
        }
        val appTitle = TextView(this).apply {
            text = "FamLoop"
            textSize = 22f
            setTypeface(null, Typeface.BOLD)
            setTextColor(Color.parseColor("#0F172A"))
        }
        val appSubtitle = TextView(this).apply {
            text = "Spatial Tracking & Family Protection"
            textSize = 12f
            setTextColor(Color.parseColor("#64748B"))
        }
        headerTextLayout.addView(appTitle)
        headerTextLayout.addView(appSubtitle)
        headerLayout.addView(headerTextLayout)
        root.addView(headerLayout)

        // Main Title & Prompt
        val promptTitle = TextView(this).apply {
            text = "Siapa yang menggunakan HP ini?"
            textSize = 24f
            setTypeface(null, Typeface.BOLD)
            setTextColor(Color.parseColor("#0F172A"))
            setPadding(0, 0, 0, 8)
        }
        root.addView(promptTitle)

        val promptDesc = TextView(this).apply {
            text = "Pilih profil untuk mengonfigurasi pelacakan dan izin pada perangkat ini."
            textSize = 14f
            setTextColor(Color.parseColor("#64748B"))
            setLineSpacing(4f, 1f)
            setPadding(0, 0, 0, 48)
        }
        root.addView(promptDesc)

        // Card 1: Parent / Guardian
        parentCard = createCard(
            title = "Orang Tua / Kepala Keluarga",
            badgeText = "ADMIN & FULL ACCESS",
            badgeBg = "#EFF6FF",
            badgeColor = "#2563EB",
            iconText = "🛡️",
            desc = "Akses penuh peta live telemetri, zona aman geofence, dan pemindai QR untuk menambah perangkat.",
            bullets = listOf(
                "Peta live telemetri real-time anggota keluarga",
                "Notifikasi geofence zona aman & pemantau SOS",
                "Pemindai QR kamera untuk menambah HP anak"
            ),
            isSelected = selectedRole == "parent"
        ) {
            selectedRole = "parent"
            updateCardSelections()
        }
        root.addView(parentCard)

        // Spacer
        val spacer = TextView(this).apply { height = 36 }
        root.addView(spacer)

        // Card 2: Child / Family Member
        memberCard = createCard(
            title = "Anak / Anggota Keluarga",
            badgeText = "SAFE DEVICE NODE",
            badgeBg = "#ECFDF5",
            badgeColor = "#059669",
            iconText = "🧒",
            desc = "Pelacakan otomatis hemat baterai dan pembuat kode QR hardware instan tanpa email/password.",
            bullets = listOf(
                "Pelacakan otomatis hemat baterai (<2.4% / hari)",
                "Pembuat kode QR hardware instan tanpa akun",
                "Privasi & keamanan data keluarga terlindungi"
            ),
            isSelected = selectedRole == "member"
        ) {
            selectedRole = "member"
            updateCardSelections()
        }
        root.addView(memberCard)

        // Spacer
        val spacerBottom = TextView(this).apply { height = 48 }
        root.addView(spacerBottom)

        // Continue Button
        continueBtn = Button(this).apply {
            text = "Lanjutkan sebagai Orang Tua"
            setTextColor(Color.WHITE)
            textSize = 15f
            setTypeface(null, Typeface.BOLD)
            val btnBg = GradientDrawable().apply {
                setColor(Color.parseColor("#2563EB"))
                cornerRadius = 24f
            }
            background = btnBg
            setPadding(32, 28, 32, 28)
            setOnClickListener {
                handleContinue()
            }
            layoutParams = LinearLayout.LayoutParams(
                LinearLayout.LayoutParams.MATCH_PARENT,
                LinearLayout.LayoutParams.WRAP_CONTENT
            )
        }
        root.addView(continueBtn)

        // Footnote
        val footnote = TextView(this).apply {
            text = "🔒 Privacy by Design: End-to-End Encrypted Loops\nPT FUJIWARA CREATIVE SOLUSINDO"
            textSize = 11f
            setTextColor(Color.parseColor("#94A3B8"))
            gravity = Gravity.CENTER
            setPadding(0, 32, 0, 0)
        }
        root.addView(footnote)

        scrollView.addView(root)
        setContentView(scrollView)
    }

    private fun createCard(
        title: String,
        badgeText: String,
        badgeBg: String,
        badgeColor: String,
        iconText: String,
        desc: String,
        bullets: List<String>,
        isSelected: Boolean,
        onClick: () -> Unit
    ): LinearLayout {
        val card = LinearLayout(this).apply {
            orientation = LinearLayout.VERTICAL
            setPadding(40, 40, 40, 40)
            applyCardBackground(this, isSelected)
            setOnClickListener { onClick() }
        }

        val topRow = LinearLayout(this).apply {
            orientation = LinearLayout.HORIZONTAL
            gravity = Gravity.CENTER_VERTICAL
            setPadding(0, 0, 0, 20)
        }

        val iconBadge = TextView(this).apply {
            text = iconText
            textSize = 20f
            gravity = Gravity.CENTER
            val bg = GradientDrawable().apply {
                setColor(Color.parseColor(badgeBg))
                cornerRadius = 20f
            }
            background = bg
            layoutParams = LinearLayout.LayoutParams(80, 80)
        }
        topRow.addView(iconBadge)

        val titleContainer = LinearLayout(this).apply {
            orientation = LinearLayout.VERTICAL
            setPadding(20, 0, 0, 0)
            layoutParams = LinearLayout.LayoutParams(0, LinearLayout.LayoutParams.WRAP_CONTENT, 1f)
        }
        val tView = TextView(this).apply {
            text = title
            textSize = 17f
            setTypeface(null, Typeface.BOLD)
            setTextColor(Color.parseColor("#0F172A"))
        }
        val bView = TextView(this).apply {
            text = badgeText
            textSize = 10f
            setTypeface(null, Typeface.BOLD)
            setTextColor(Color.parseColor(badgeColor))
            setPadding(0, 4, 0, 0)
        }
        titleContainer.addView(tView)
        titleContainer.addView(bView)
        topRow.addView(titleContainer)

        card.addView(topRow)

        val descView = TextView(this).apply {
            text = desc
            textSize = 13f
            setTextColor(Color.parseColor("#64748B"))
            setLineSpacing(3f, 1f)
            setPadding(0, 0, 0, 20)
        }
        card.addView(descView)

        // Bullets
        bullets.forEach { bullet ->
            val bRow = TextView(this).apply {
                text = "✓  $bullet"
                textSize = 12f
                setTextColor(Color.parseColor("#334155"))
                setPadding(0, 4, 0, 4)
            }
            card.addView(bRow)
        }

        return card
    }

    private fun applyCardBackground(layout: LinearLayout, isSelected: Boolean) {
        val bg = GradientDrawable().apply {
            setColor(Color.WHITE)
            cornerRadius = 32f
            if (isSelected) {
                setStroke(4, Color.parseColor("#2563EB"))
            } else {
                setStroke(2, Color.parseColor("#E2E8F0"))
            }
        }
        layout.background = bg
    }

    private fun updateCardSelections() {
        applyCardBackground(parentCard, selectedRole == "parent")
        applyCardBackground(memberCard, selectedRole == "member")

        if (selectedRole == "parent") {
            continueBtn.text = "Lanjutkan sebagai Orang Tua"
            val bg = GradientDrawable().apply {
                setColor(Color.parseColor("#2563EB"))
                cornerRadius = 24f
            }
            continueBtn.background = bg
        } else {
            continueBtn.text = "Lanjutkan sebagai Anak"
            val bg = GradientDrawable().apply {
                setColor(Color.parseColor("#059669"))
                cornerRadius = 24f
            }
            continueBtn.background = bg
        }
    }

    private fun handleContinue() {
        session.role = selectedRole
        if (selectedRole == "parent") {
            if (session.isLoggedIn && !session.groupId.isNullOrEmpty()) {
                val intent = Intent(this, LiveMapActivity::class.java).apply {
                    putExtra("EXTRA_GROUP_ID", session.groupId)
                    putExtra("EXTRA_GROUP_NAME", session.groupName)
                }
                startActivity(intent)
            } else {
                startActivity(Intent(this, ParentAuthActivity::class.java))
            }
        } else {
            startActivity(Intent(this, MemberQrActivity::class.java))
        }
        finish()
    }
}
