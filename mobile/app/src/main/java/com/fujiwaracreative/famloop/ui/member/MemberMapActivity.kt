package com.fujiwaracreative.famloop.ui.member

import android.content.Intent
import android.graphics.Color
import android.graphics.Typeface
import android.graphics.drawable.GradientDrawable
import android.os.Bundle
import android.view.Gravity
import android.widget.Button
import android.widget.FrameLayout
import android.widget.LinearLayout
import android.widget.Toast
import androidx.appcompat.app.AppCompatActivity
import com.fujiwaracreative.famloop.data.local.SessionManager
import com.fujiwaracreative.famloop.data.remote.NetworkClient
import com.fujiwaracreative.famloop.service.LocationForegroundService
import com.fujiwaracreative.famloop.ui.MainActivity
import com.fujiwaracreative.famloop.ui.common.LeafletMapView
import kotlinx.coroutines.CoroutineScope
import kotlinx.coroutines.Dispatchers
import kotlinx.coroutines.launch
import okhttp3.Response
import okhttp3.sse.EventSource
import okhttp3.sse.EventSourceListener
import org.json.JSONObject

class MemberMapActivity : AppCompatActivity() {

    private lateinit var mapView: LeafletMapView
    private var eventSource: EventSource? = null
    private var currentGroupId: String = ""
    private var currentDeviceId: String = ""

    override fun onCreate(savedInstanceState: Bundle?) {
        super.onCreate(savedInstanceState)
        val session = SessionManager.getInstance(this)
        session.token?.let { NetworkClient.authToken = it }

        currentGroupId = intent.getStringExtra("EXTRA_GROUP_ID") ?: (session.groupId ?: "")
        currentDeviceId = intent.getStringExtra("EXTRA_DEVICE_ID") ?: ""

        val root = FrameLayout(this)
        mapView = LeafletMapView(this)
        root.addView(mapView, FrameLayout.LayoutParams(FrameLayout.LayoutParams.MATCH_PARENT, FrameLayout.LayoutParams.MATCH_PARENT))

        // Top Action Bar
        val topBar = LinearLayout(this).apply {
            orientation = LinearLayout.HORIZONTAL
            gravity = Gravity.CENTER_VERTICAL
            setPadding(32, 48, 32, 24)
            layoutParams = FrameLayout.LayoutParams(
                FrameLayout.LayoutParams.MATCH_PARENT,
                FrameLayout.LayoutParams.WRAP_CONTENT
            ).apply {
                gravity = Gravity.TOP
            }
        }

        val qrButton = Button(this).apply {
            text = "📱 QR Perangkat"
            val bg = GradientDrawable().apply {
                setColor(Color.parseColor("#EEFFFFFF"))
                cornerRadius = 20f
                setStroke(2, Color.parseColor("#E2E8F0"))
            }
            background = bg
            setTextColor(Color.parseColor("#0F172A"))
            textSize = 13f
            setTypeface(null, Typeface.BOLD)
            setPadding(32, 20, 32, 20)
            setOnClickListener {
                startActivity(Intent(this@MemberMapActivity, MemberQrActivity::class.java))
                finish()
            }
        }
        topBar.addView(qrButton)

        val spacer = android.view.View(this).apply {
            layoutParams = LinearLayout.LayoutParams(0, 1, 1f)
        }
        topBar.addView(spacer)

        val switchBtn = Button(this).apply {
            text = "Ganti Peran"
            val bg = GradientDrawable().apply {
                setColor(Color.parseColor("#EEFFFFFF"))
                cornerRadius = 20f
                setStroke(2, Color.parseColor("#E2E8F0"))
            }
            background = bg
            setTextColor(Color.parseColor("#475569"))
            textSize = 12f
            setPadding(28, 20, 28, 20)
            setOnClickListener {
                val intent = Intent(this@MemberMapActivity, MainActivity::class.java).apply {
                    putExtra("EXTRA_FORCE_CHOOSER", true)
                }
                startActivity(intent)
                finish()
            }
        }
        topBar.addView(switchBtn)

        root.addView(topBar)
        setContentView(root)

        // Otomatis nyalakan background foreground service tracking bila sudah dipasangkan
        if (currentDeviceId.isNotEmpty() && currentGroupId.isNotEmpty()) {
            LocationForegroundService.start(this, currentDeviceId, currentGroupId)
        }

        if (currentGroupId.isNotEmpty()) {
            loadInitialLocations()
            startRealtimeStreaming()
        }
    }

    private fun loadInitialLocations() {
        CoroutineScope(Dispatchers.IO).launch {
            try {
                val resp = NetworkClient.apiService.getLatestLocations(currentGroupId)
                if (resp.isSuccessful) {
                    val members = resp.body()?.data ?: emptyList()
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
                    Toast.makeText(this@MemberMapActivity, "Info: ${e.message}", Toast.LENGTH_SHORT).show()
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
                        "MemberAvatarUpdated" -> {
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
