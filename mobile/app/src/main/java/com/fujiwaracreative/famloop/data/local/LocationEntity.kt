package com.fujiwaracreative.famloop.data.local

import androidx.room.Entity
import androidx.room.Index
import androidx.room.PrimaryKey

@Entity(
    tableName = "offline_locations",
    indices = [
        Index(value = ["deviceId", "recordedAt"], unique = true),
        Index(value = ["isSynced"])
    ]
)
data class LocationEntity(
    @PrimaryKey(autoGenerate = true)
    val id: Long = 0,
    val deviceId: String,
    val groupId: String,
    val latitude: Double,
    val longitude: Double,
    val accuracy: Float,
    val speed: Float,
    val heading: Float,
    val batteryLevel: Int,
    val recordedAt: Long,
    val isSynced: Boolean = false
)
