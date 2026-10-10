package com.fujiwaracreative.famloop.data.local

import androidx.room.Dao
import androidx.room.Insert
import androidx.room.OnConflictStrategy
import androidx.room.Query

@Dao
interface LocationDao {

    @Insert(onConflict = OnConflictStrategy.IGNORE)
    suspend fun insertPoint(point: LocationEntity): Long

    @Query("SELECT * FROM offline_locations WHERE isSynced = 0 ORDER BY recordedAt ASC LIMIT :limit")
    suspend fun getUnsyncedPoints(limit: Int = 100): List<LocationEntity>

    @Query("UPDATE offline_locations SET isSynced = 1 WHERE id IN (:ids)")
    suspend fun markAsSynced(ids: List<Long>)

    @Query("DELETE FROM offline_locations WHERE isSynced = 1 AND recordedAt < :olderThanTimestamp")
    suspend fun deleteOldSyncedPoints(olderThanTimestamp: Long)

    @Query("SELECT COUNT(*) FROM offline_locations WHERE isSynced = 0")
    suspend fun getUnsyncedCount(): Int
}
