package com.fujiwaracreative.famloop.service

import android.annotation.SuppressLint
import android.app.Notification
import android.app.NotificationChannel
import android.app.NotificationManager
import android.app.Service
import android.content.Context
import android.content.Intent
import android.os.Build
import android.os.IBinder
import android.os.Looper
import android.util.Log
import androidx.core.app.NotificationCompat
import com.fujiwaracreative.famloop.data.local.AppDatabase
import com.fujiwaracreative.famloop.data.local.LocationEntity
import com.fujiwaracreative.famloop.data.remote.NetworkClient
import com.fujiwaracreative.famloop.data.remote.PingRequest
import com.fujiwaracreative.famloop.utils.DeviceUtils
import com.fujiwaracreative.famloop.worker.LocationSyncWorker
import com.google.android.gms.location.*
import kotlinx.coroutines.CoroutineScope
import kotlinx.coroutines.Dispatchers
import kotlinx.coroutines.SupervisorJob
import kotlinx.coroutines.launch

class LocationForegroundService : Service() {

    private val serviceScope = CoroutineScope(Dispatchers.IO + SupervisorJob())
    private lateinit var fusedLocationClient: FusedLocationProviderClient
    private lateinit var locationCallback: LocationCallback
    private lateinit var database: AppDatabase

    private var currentDeviceId: String = ""
    private var currentGroupId: String = ""

    override fun onCreate() {
        super.onCreate()
        database = AppDatabase.getInstance(this)
        fusedLocationClient = LocationServices.getFusedLocationProviderClient(this)

        createNotificationChannel()
        startForeground(NOTIFICATION_ID, buildForegroundNotification("FamLoop aktif memantau lokasi keluarga"))

        setupLocationUpdates()
    }

    override fun onStartCommand(intent: Intent?, flags: Int, startId: Int): Int {
        intent?.let {
            currentDeviceId = it.getStringExtra(EXTRA_DEVICE_ID) ?: currentDeviceId
            currentGroupId = it.getStringExtra(EXTRA_GROUP_ID) ?: currentGroupId
        }
        return START_STICKY
    }

    @SuppressLint("MissingPermission")
    private fun setupLocationUpdates() {
        val locationRequest = LocationRequest.Builder(Priority.PRIORITY_HIGH_ACCURACY, 15000L).apply {
            setMinUpdateIntervalMillis(10000L)
            setWaitForAccurateLocation(false)
        }.build()

        locationCallback = object : LocationCallback() {
            override fun onLocationResult(result: LocationResult) {
                val loc = result.lastLocation ?: return
                val now = System.currentTimeMillis()
                val battery = DeviceUtils.getBatteryLevel(this@LocationForegroundService)

                serviceScope.launch {
                    val localId = database.locationDao().insertPoint(
                        LocationEntity(
                            deviceId = currentDeviceId,
                            groupId = currentGroupId,
                            latitude = loc.latitude,
                            longitude = loc.longitude,
                            accuracy = loc.accuracy,
                            speed = loc.speed,
                            heading = loc.bearing,
                            batteryLevel = battery,
                            recordedAt = now,
                            isSynced = false
                        )
                    )

                    // Coba kirim langsung jika ada koneksi
                    try {
                        val resp = NetworkClient.apiService.sendPing(
                            PingRequest(
                                device_id = currentDeviceId,
                                group_id = currentGroupId,
                                latitude = loc.latitude,
                                longitude = loc.longitude,
                                accuracy = loc.accuracy,
                                speed = loc.speed,
                                heading = loc.bearing,
                                battery_level = battery,
                                recorded_at = now
                            )
                        )
                        if (resp.isSuccessful) {
                            database.locationDao().markAsSynced(listOf(localId))
                        } else {
                            // Gagal kirim ke server, jadwalkan SyncWorker
                            LocationSyncWorker.enqueue(this@LocationForegroundService)
                        }
                    } catch (e: Exception) {
                        Log.d("LocationService", "Offline, data tersimpan di Room SQLite: ${e.message}")
                        LocationSyncWorker.enqueue(this@LocationForegroundService)
                    }
                }
            }
        }

        fusedLocationClient.requestLocationUpdates(locationRequest, locationCallback, Looper.getMainLooper())
    }

    private fun createNotificationChannel() {
        if (Build.VERSION.SDK_INT >= Build.VERSION_CODES.O) {
            val channel = NotificationChannel(
                CHANNEL_ID,
                "FamLoop Live Location",
                NotificationManager.IMPORTANCE_LOW
            ).apply {
                description = "Layanan pemantauan lokasi latar belakang FamLoop"
            }
            val manager = getSystemService(NotificationManager::class.java)
            manager?.createNotificationChannel(channel)
        }
    }

    private fun buildForegroundNotification(contentText: String): Notification {
        return NotificationCompat.Builder(this, CHANNEL_ID)
            .setContentTitle("FamLoop Family Tracking")
            .setContentText(contentText)
            .setSmallIcon(android.R.drawable.ic_menu_mylocation)
            .setOngoing(true)
            .setPriority(NotificationCompat.PRIORITY_LOW)
            .build()
    }

    override fun onDestroy() {
        super.onDestroy()
        fusedLocationClient.removeLocationUpdates(locationCallback)
    }

    override fun onBind(intent: Intent?): IBinder? = null

    companion object {
        const val NOTIFICATION_ID = 1001
        const val CHANNEL_ID = "famloop_location_channel"
        const val EXTRA_DEVICE_ID = "extra_device_id"
        const val EXTRA_GROUP_ID = "extra_group_id"

        fun start(context: Context, deviceId: String, groupId: String) {
            val intent = Intent(context, LocationForegroundService::class.java).apply {
                putExtra(EXTRA_DEVICE_ID, deviceId)
                putExtra(EXTRA_GROUP_ID, groupId)
            }
            if (Build.VERSION.SDK_INT >= Build.VERSION_CODES.O) {
                context.startForegroundService(intent)
            } else {
                context.startService(intent)
            }
        }
    }
}
