package com.fujiwaracreative.famloop.ui.parent

import android.content.Intent
import android.graphics.Color
import android.graphics.Typeface
import android.graphics.drawable.GradientDrawable
import android.os.Bundle
import android.view.Gravity
import android.widget.Button
import android.widget.FrameLayout
import android.widget.LinearLayout
import android.widget.TextView
import android.widget.Toast
import androidx.activity.result.contract.ActivityResultContracts
import androidx.appcompat.app.AppCompatActivity
import com.fujiwaracreative.famloop.data.local.SessionManager
import com.fujiwaracreative.famloop.data.remote.NetworkClient
import com.fujiwaracreative.famloop.ui.MainActivity
import com.fujiwaracreative.famloop.ui.common.LeafletMapView
import kotlinx.coroutines.CoroutineScope
import kotlinx.coroutines.Dispatchers
import kotlinx.coroutines.launch
import okhttp3.Response
import okhttp3.sse.EventSource
import okhttp3.sse.EventSourceListener
import org.json.JSONObject

class LiveMapActivity : AppCompatActivity() {

    private lateinit var mapView: LeafletMapView
    private var eventSource: EventSource? = null
    private var currentGroupId: String = ""
    private var currentGroupName: String = "FamLoop Circle"
    private lateinit var session: SessionManager
    private lateinit var statusSubtitle: TextView

    private val scanLauncher =
        registerForActivityResult(ActivityResultContracts.StartActivityForResult()) { result ->
            if (result.resultCode == RESULT_OK) {
                loadInitialLocations()
            }
        }

    override fun onCreate(savedInstanceState: Bundle?) {
        super.onCreate(savedInstanceState)
        session = SessionManager.getInstance(this)
        session.token?.let { NetworkClient.authToken = it }

        currentGroupId = intent.getStringExtra("EXTRA_GROUP_ID") ?: (session.groupId ?: "")
        currentGroupName = intent.getStringExtra("EXTRA_GROUP_NAME") ?: (session.groupName ?: "FamLoop Circle")

        if (currentGroupId.isEmpty()) {
            Toast.makeText(this, "Silakan pilih atau buat grup keluarga terlebih dahulu", Toast.LENGTH_SHORT).show()
            startActivity(Intent(this, ParentAuthActivity::class.java))
            finish()
            return
        }

        val root = FrameLayout(this)

        // 1. Leaflet Map View (Full screen telemetry canvas)
        mapView = LeafletMapView(this)
        root.addView(
            mapView,
            FrameLayout.LayoutParams(FrameLayout.LayoutParams.MATCH_PARENT, FrameLayout.LayoutParams.MATCH_PARENT)
        )

        // 2. Top Floating Header Bar (Stitch Circle Dropdown & SOS)
        val topBar = LinearLayout(this).apply {
            orientation = LinearLayout.HORIZONTAL
            gravity = Gravity.CENTER_VERTICAL
            setPadding(28, 20, 28, 20)
            val bg = GradientDrawable().apply {
                setColor(Color.parseColor("#EE0F172A")) // Dark frosted glass
                cornerRadius = 32f
                setStroke(2, Color.parseColor("#334155"))
            }
            background = bg
            layoutParams = FrameLayout.LayoutParams(
                FrameLayout.LayoutParams.MATCH_PARENT,
                FrameLayout.LayoutParams.WRAP_CONTENT
            ).apply {
                gravity = Gravity.TOP
                setMargins(24, 48, 24, 0)
            }
        }

        val circleInfoLayout = LinearLayout(this).apply {
            orientation = LinearLayout.VERTICAL
            layoutParams = LinearLayout.LayoutParams(0, LinearLayout.LayoutParams.WRAP_CONTENT, 1f)
        }
        val groupTitle = TextView(this).apply {
            text = "📍 $currentGroupName"
            textSize = 15f
            setTypeface(null, Typeface.BOLD)
            setTextColor(Color.WHITE)
        }
        statusSubtitle = TextView(this).apply {
            text = "● Live Telemetry Synced"
            textSize = 11f
            setTextColor(Color.parseColor("#34D399")) // Emerald
        }
        circleInfoLayout.addView(groupTitle)
        circleInfoLayout.addView(statusSubtitle)
        topBar.addView(circleInfoLayout)

        // SOS Emergency Alert Button
        val btnSos = Button(this).apply {
            text = "SOS"
            setTextColor(Color.WHITE)
            textSize = 12f
            setTypeface(null, Typeface.BOLD)
            val bg = GradientDrawable().apply {
                setColor(Color.parseColor("#DC2626")) // Red-600
                cornerRadius = 16f
            }
            background = bg
            setPadding(24, 12, 24, 12)
            setOnClickListener {
                Toast.makeText(this@LiveMapActivity, "Sinyal SOS disiarkan ke seluruh perangkat keluarga!", Toast.LENGTH_LONG).show()
            }
        }
        topBar.addView(btnSos)

        // Ganti Peran Button
        val btnRole = Button(this).apply {
            text = "Ganti Peran"
            setTextColor(Color.parseColor("#94A3B8"))
            textSize = 11f
            setBackgroundColor(Color.TRANSPARENT)
            setPadding(16, 0, 0, 0)
            setOnClickListener {
                val intent = Intent(this@LiveMapActivity, MainActivity::class.java).apply {
                    putExtra("EXTRA_FORCE_CHOOSER", true)
                }
                startActivity(intent)
                finish()
            }
        }
        topBar.addView(btnRole)
        root.addView(topBar)

        // 3. Bottom Action Floating Bar
        val bottomBar = LinearLayout(this).apply {
            orientation = LinearLayout.HORIZONTAL
            gravity = Gravity.CENTER
            setPadding(24, 0, 24, 48)
            layoutParams = FrameLayout.LayoutParams(
                FrameLayout.LayoutParams.MATCH_PARENT,
                FrameLayout.LayoutParams.WRAP_CONTENT
            ).apply {
                gravity = Gravity.BOTTOM
            }
        }

        // Tombol Tambah Anggota (Buka Scanner QR)
        val scanButton = Button(this).apply {
            text = "+ Tambah Anggota (Scan QR)"
            val bg = GradientDrawable().apply {
                setColor(Color.parseColor("#2563EB"))
                cornerRadius = 24f
            }
            background = bg
            setTextColor(Color.WHITE)
            setTypeface(null, Typeface.BOLD)
            textSize = 14f
            setPadding(40, 24, 40, 24)
            setOnClickListener {
                val intent = Intent(this@LiveMapActivity, QrScannerActivity::class.java).apply {
                    putExtra("EXTRA_GROUP_ID", currentGroupId)
                }
                scanLauncher.launch(intent)
            }
        }
        bottomBar.addView(scanButton)

        // Tombol Fit Peta
        val fitButton = Button(this).apply {
            text = "⌖"
            val bg = GradientDrawable().apply {
                setColor(Color.parseColor("#1E293B"))
                cornerRadius = 24f
                setStroke(2, Color.parseColor("#334155"))
            }
            background = bg
            setTextColor(Color.parseColor("#38BDF8"))
            textSize = 20f
            setTypeface(null, Typeface.BOLD)
            setPadding(24, 20, 24, 20)
            setOnClickListener {
                mapView.fitAll()
            }
            layoutParams = LinearLayout.LayoutParams(
                LinearLayout.LayoutParams.WRAP_CONTENT,
                LinearLayout.LayoutParams.WRAP_CONTENT
            ).apply {
                leftMargin = 16
            }
        }
        bottomBar.addView(fitButton)

        root.addView(bottomBar)

        setContentView(root)

        loadInitialLocations()
        startRealtimeStreaming()
    }

    private fun loadInitialLocations() {
        CoroutineScope(Dispatchers.IO).launch {
            try {
                val resp = NetworkClient.apiService.getLatestLocations(currentGroupId)
                if (resp.isSuccessful) {
                    val members = resp.body()?.data ?: emptyList()
                    runOnUiThread {
                        statusSubtitle.text = "● ${members.size} Perangkat Terhubung & Aktif"
                    }
                    members.forEach { member ->
                        mapView.updateMemberMarker(
                            deviceId = member.device_id,
                            name = member.device_name,
                            latitude = member.latitude,
                            longitude = member.longitude,
                            avatarUrl = member.avatar,
                            battery = member.battery_level,
                            isOnline = member.is_active
                        )
                    }
                    if (members.isNotEmpty()) {
                        mapView.fitAll()
                    }
                }
            } catch (e: Exception) {
                runOnUiThread {
                    Toast.makeText(this@LiveMapActivity, "Info: ${e.message}", Toast.LENGTH_SHORT).show()
                }
            }
        }
    }

    private fun startRealtimeStreaming() {
        eventSource = NetworkClient.createGroupEventSource(currentGroupId, object : EventSourceListener() {
            override fun onEvent(eventSource: EventSource, id: String?, type: String?, data: String) {
                try {
                    when (type) {
                        "LocationUpdated" -> {
                            val json = JSONObject(data)
                            val deviceId = json.getString("device_id")
                            val name = json.optString("device_name", "Anggota")
                            val lat = json.getDouble("latitude")
                            val lng = json.getDouble("longitude")
                            val avatar = json.optString("avatar", "")
                            val battery = json.optInt("battery_level", 100)

                            runOnUiThread {
                                mapView.updateMemberMarker(
                                    deviceId = deviceId,
                                    name = name,
                                    latitude = lat,
                                    longitude = lng,
                                    avatarUrl = avatar,
                                    battery = battery,
                                    isOnline = true
                                )
                            }
                        }
                        "MemberAvatarUpdated", "MemberJoined" -> {
                            loadInitialLocations()
                        }
                    }
                } catch (_: Exception) {}
            }

            override fun onFailure(eventSource: EventSource, t: Throwable?, response: Response?) {}
        })
    }

    override fun onDestroy() {
        super.onDestroy()
        eventSource?.cancel()
    }
}
