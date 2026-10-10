package com.fujiwaracreative.famloop.worker

import android.content.Context
import androidx.work.*
import com.fujiwaracreative.famloop.data.local.AppDatabase
import com.fujiwaracreative.famloop.data.remote.NetworkClient
import com.fujiwaracreative.famloop.data.remote.PingRequest
import com.fujiwaracreative.famloop.data.remote.SyncBatchRequest

class LocationSyncWorker(
    context: Context,
    workerParams: WorkerParameters
) : CoroutineWorker(context, workerParams) {

    override suspend fun doWork(): Result {
        val database = AppDatabase.getInstance(applicationContext)
        val unsyncedList = database.locationDao().getUnsyncedPoints(limit = 100)

        if (unsyncedList.isEmpty()) {
            return Result.success()
        }

        val groupId = unsyncedList.first().groupId
        val pingRequests = unsyncedList.map { entity ->
            PingRequest(
                device_id = entity.deviceId,
                group_id = entity.groupId,
                latitude = entity.latitude,
                longitude = entity.longitude,
                accuracy = entity.accuracy,
                speed = entity.speed,
                heading = entity.heading,
                battery_level = entity.batteryLevel,
                recorded_at = entity.recordedAt
            )
        }

        return try {
            val response = NetworkClient.apiService.sendSyncBatch(
                SyncBatchRequest(
                    group_id = groupId,
                    points = pingRequests
                )
            )

            if (response.isSuccessful) {
                val syncedIds = unsyncedList.map { it.id }
                database.locationDao().markAsSynced(syncedIds)

                // Hapus data tersinkron yang lebih lama dari 7 hari agar SQLite lokal tetap ramping
                val sevenDaysAgo = System.currentTimeMillis() - (7 * 24 * 60 * 60 * 1000L)
                database.locationDao().deleteOldSyncedPoints(sevenDaysAgo)

                // Jika masih ada sisa antrean yang belum terkirim, jadwalkan kembali
                if (database.locationDao().getUnsyncedCount() > 0) {
                    Result.retry()
                } else {
                    Result.success()
                }
            } else {
                Result.retry()
            }
        } catch (e: Exception) {
            Result.retry()
        }
    }

    companion object {
        private const val UNIQUE_WORK_NAME = "FamLoopLocationSyncWork"

        fun enqueue(context: Context) {
            val constraints = Constraints.Builder()
                .setRequiredNetworkType(NetworkType.CONNECTED)
                .build()

            val syncRequest = OneTimeWorkRequestBuilder<LocationSyncWorker>()
                .setConstraints(constraints)
                .build()

            WorkManager.getInstance(context).enqueueUniqueWork(
                UNIQUE_WORK_NAME,
                ExistingWorkPolicy.KEEP,
                syncRequest
            )
        }
    }
}
