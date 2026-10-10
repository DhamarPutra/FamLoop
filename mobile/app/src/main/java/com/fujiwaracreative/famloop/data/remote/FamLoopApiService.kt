package com.fujiwaracreative.famloop.data.remote

import okhttp3.MultipartBody
import retrofit2.Response
import retrofit2.http.*

data class PingRequest(
    val device_id: String,
    val group_id: String,
    val latitude: Double,
    val longitude: Double,
    val accuracy: Float,
    val speed: Float,
    val heading: Float,
    val battery_level: Int,
    val recorded_at: Long
)

data class SyncBatchRequest(
    val group_id: String,
    val points: List<PingRequest>
)

data class ApiResponse<T>(
    val ok: Boolean? = null,
    val message: String? = null,
    val data: T? = null
)

data class MemberLocationResponse(
    val device_id: String,
    val device_uuid: String,
    val device_name: String,
    val avatar: String,
    val battery_level: Int,
    val is_active: Boolean,
    val latitude: Double,
    val longitude: Double,
    val accuracy: Float,
    val speed: Float,
    val heading: Float,
    val recorded_at: Long,
    val last_seen_at: String
)

data class PairDeviceRequest(
    val device_uuid: String,
    val device_name: String,
    val avatar: String? = null,
    val ipv4: String? = null,
    val ipv6: String? = null,
    val mac_address: String? = null
)

interface FamLoopApiService {

    @POST("api/v1/tracking/ping")
    suspend fun sendPing(@Body request: PingRequest): Response<ApiResponse<Any>>

    @POST("api/v1/tracking/sync/batch")
    suspend fun sendSyncBatch(@Body request: SyncBatchRequest): Response<ApiResponse<Any>>

    @GET("api/v1/groups/{id}/locations/latest")
    suspend fun getLatestLocations(@Path("id") groupId: String): Response<ApiResponse<List<MemberLocationResponse>>>

    @POST("api/v1/groups/{id}/pair-device")
    suspend fun pairDevice(
        @Path("id") groupId: String,
        @Body request: PairDeviceRequest
    ): Response<ApiResponse<Any>>

    @Multipart
    @POST("api/v1/devices/{id}/avatar")
    suspend fun uploadDeviceAvatar(
        @Path("id") deviceId: String,
        @Part avatar: MultipartBody.Part,
        @Query("group_id") groupId: String
    ): Response<ApiResponse<Any>>

    @Multipart
    @POST("api/v1/auth/avatar")
    suspend fun uploadUserAvatar(
        @Part avatar: MultipartBody.Part
    ): Response<ApiResponse<Any>>
}
