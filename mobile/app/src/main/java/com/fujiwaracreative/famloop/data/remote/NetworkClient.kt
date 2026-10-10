package com.fujiwaracreative.famloop.data.remote

import okhttp3.Interceptor
import okhttp3.OkHttpClient
import okhttp3.Request
import okhttp3.logging.HttpLoggingInterceptor
import okhttp3.sse.EventSource
import okhttp3.sse.EventSourceListener
import okhttp3.sse.EventSources
import retrofit2.Retrofit
import retrofit2.converter.gson.GsonConverterFactory
import java.util.concurrent.TimeUnit

object NetworkClient {

    // Production API endpoint FamLoop
    var BASE_URL = "https://famloop.my.id/"

    @Volatile
    var authToken: String? = null

    private val authInterceptor = Interceptor { chain ->
        val original = chain.request()
        val builder = original.newBuilder()
        authToken?.let { token ->
            if (token.isNotEmpty()) {
                builder.header("Authorization", "Bearer $token")
            }
        }
        chain.proceed(builder.build())
    }

    private val okHttpClient by lazy {
        OkHttpClient.Builder()
            .connectTimeout(15, TimeUnit.SECONDS)
            .readTimeout(0, TimeUnit.SECONDS) // Read timeout 0 untuk SSE persistent connection
            .writeTimeout(30, TimeUnit.SECONDS)
            .addInterceptor(authInterceptor)
            .addInterceptor(HttpLoggingInterceptor().apply {
                level = HttpLoggingInterceptor.Level.BODY
            })
            .build()
    }

    val apiService: FamLoopApiService by lazy {
        Retrofit.Builder()
            .baseUrl(BASE_URL)
            .client(okHttpClient)
            .addConverterFactory(GsonConverterFactory.create())
            .build()
            .create(FamLoopApiService::class.java)
    }

    // Membuka stream SSE Server-Sent Events real-time ke endpoint /api/v1/groups/:id/sse
    fun createGroupEventSource(groupId: String, listener: EventSourceListener): EventSource {
        val sseUrl = "${BASE_URL}api/v1/groups/$groupId/sse"
        val requestBuilder = Request.Builder()
            .url(sseUrl)
            .header("Accept", "text/event-stream")

        authToken?.let { token ->
            if (token.isNotEmpty()) {
                requestBuilder.header("Authorization", "Bearer $token")
            }
        }

        val factory = EventSources.createFactory(okHttpClient)
        return factory.newEventSource(requestBuilder.build(), listener)
    }
}
