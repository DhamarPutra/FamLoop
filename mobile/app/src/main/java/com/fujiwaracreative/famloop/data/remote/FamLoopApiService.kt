package com.fujiwaracreative.famloop.data.remote

import okhttp3.MultipartBody
import retrofit2.Response
import retrofit2.http.*

data class RegisterRequest(
    val username: String,
    val email: String,
    val phone_number: String,
    val password: String
)

data class LoginRequest(
    val username_or_email: String,
    val password: String
)

data class UserData(
    val id: String,
    val username: String,
    val email: String,
    val phone_number: String,
    val avatar: String? = null
)

data class AuthResponseData(
    val token: String,
    val user: UserData
)

data class CreateGroupRequest(
    val name: String
)

data class GroupData(
    val id: String,
    val name: String,
    val invite_code: String? = null,
    val created_at: String? = null
)

data class MemberDeviceData(
    val id: String,
    val group_id: String,
    val device_uuid: String,
    val device_name: String,
    val avatar: String? = null,
    val battery_level: Int = 100,
    val is_active: Boolean = true
)

data class GroupDetailsData(
    val group: GroupData,
    val members: List<MemberDeviceData>
)

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
    val data: T? = null,
    val error: Any? = null
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

    // Auth
    @POST("api/v1/auth/register")
    suspend fun register(@Body request: RegisterRequest): Response<ApiResponse<AuthResponseData>>

    @POST("api/v1/auth/login")
    suspend fun login(@Body request: LoginRequest): Response<ApiResponse<AuthResponseData>>

    // Groups
    @POST("api/v1/groups")
    suspend fun createGroup(@Body request: CreateGroupRequest): Response<ApiResponse<GroupData>>

    @GET("api/v1/groups")
    suspend fun getGroups(): Response<ApiResponse<List<GroupData>>>

    @GET("api/v1/groups/{id}")
    suspend fun getGroupDetails(@Path("id") groupId: String): Response<ApiResponse<GroupDetailsData>>

    // Tracking & Pings
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
