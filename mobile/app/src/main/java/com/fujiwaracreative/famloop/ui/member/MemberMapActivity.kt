package com.fujiwaracreative.famloop.ui.member

import android.content.Intent
import android.os.Bundle
import android.widget.Button
import android.widget.FrameLayout
import android.widget.Toast
import androidx.appcompat.app.AppCompatActivity
import com.fujiwaracreative.famloop.data.remote.NetworkClient
import com.fujiwaracreative.famloop.service.LocationForegroundService
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
    private var currentGroupId: String = "default-group-id"
    private var currentDeviceId: String = ""

    override fun onCreate(savedInstanceState: Bundle?) {
        super.onCreate(savedInstanceState)
        currentGroupId = intent.getStringExtra("EXTRA_GROUP_ID") ?: currentGroupId
        currentDeviceId = intent.getStringExtra("EXTRA_DEVICE_ID") ?: currentDeviceId

        val root = FrameLayout(this)
        mapView = LeafletMapView(this)
        root.addView(mapView, FrameLayout.LayoutParams(FrameLayout.LayoutParams.MATCH_PARENT, FrameLayout.LayoutParams.MATCH_PARENT))

        // Tombol QR Code
        val qrButton = Button(this).apply {
            text = "Tampilkan QR Perangkat"
            setBackgroundColor(android.graphics.Color.parseColor("#475569"))
            setTextColor(android.graphics.Color.WHITE)
            setOnClickListener {
                startActivity(Intent(this@MemberMapActivity, MemberQrActivity::class.java))
            }
        }
        val btnParams = FrameLayout.LayoutParams(
            FrameLayout.LayoutParams.WRAP_CONTENT,
            FrameLayout.LayoutParams.WRAP_CONTENT
        ).apply {
            gravity = android.view.Gravity.BOTTOM or android.view.Gravity.CENTER_HORIZONTAL
            bottomMargin = 48
        }
        root.addView(qrButton, btnParams)

        setContentView(root)

        // Otomatis nyalakan background foreground service tracking
        if (currentDeviceId.isNotEmpty() && currentGroupId.isNotEmpty()) {
            LocationForegroundService.start(this, currentDeviceId, currentGroupId)
        }

        loadInitialLocations()
        startRealtimeStreaming()
    }

    private fun loadInitialLocations() {
        CoroutineScope(Dispatchers.IO).launch {
            try {
                val resp = NetworkClient.apiService.getLatestLocations(currentGroupId)
                if (resp.isSuccessful) {
                    resp.body()?.data?.forEach { member ->
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
                    mapView.fitAll()
                }
            } catch (e: Exception) {
                runOnUiThread {
                    Toast.makeText(this@MemberMapActivity, "Gagal memuat titik lokasi: ${e.message}", Toast.LENGTH_SHORT).show()
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
