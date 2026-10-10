package com.fujiwaracreative.famloop.ui.common

import android.annotation.SuppressLint
import android.content.Context
import android.util.AttributeSet
import android.webkit.WebSettings
import android.webkit.WebView
import android.webkit.WebViewClient
import android.widget.FrameLayout

@SuppressLint("SetJavaScriptEnabled")
class LeafletMapView @JvmOverloads constructor(
    context: Context,
    attrs: AttributeSet? = null,
    defStyleAttr: Int = 0
) : FrameLayout(context, attrs, defStyleAttr) {

    private val webView: WebView = WebView(context)
    private var isMapLoaded = false
    private val pendingUpdates = mutableListOf<String>()

    init {
        addView(
            webView,
            LayoutParams(LayoutParams.MATCH_PARENT, LayoutParams.MATCH_PARENT)
        )

        webView.settings.apply {
            javaScriptEnabled = true
            domStorageEnabled = true
            cacheMode = WebSettings.LOAD_DEFAULT
            allowFileAccess = true
        }

        webView.webViewClient = object : WebViewClient() {
            override fun onPageFinished(view: WebView?, url: String?) {
                super.onPageFinished(view, url)
                isMapLoaded = true
                for (script in pendingUpdates) {
                    webView.evaluateJavascript(script, null)
                }
                pendingUpdates.clear()
            }
        }

        webView.loadUrl("file:///android_asset/map/index.html")
    }

    // Menggeser / memperbarui avatar marker anggota keluarga
    fun updateMemberMarker(
        deviceId: String,
        name: String,
        latitude: Double,
        longitude: Double,
        avatarUrl: String,
        battery: Int = 100,
        isOnline: Boolean = true
    ) {
        val safeName = name.replace("'", "\\'")
        val safeAvatar = avatarUrl.replace("'", "\\'")
        val script = "updateMemberMarker('$deviceId', '$safeName', $latitude, $longitude, '$safeAvatar', $battery, $isOnline);"

        if (isMapLoaded) {
            post { webView.evaluateJavascript(script, null) }
        } else {
            pendingUpdates.add(script)
        }
    }

    // Memusatkan peta ke anggota tertentu
    fun centerOn(latitude: Double, longitude: Double) {
        val script = "centerToMember($latitude, $longitude);"
        if (isMapLoaded) {
            post { webView.evaluateJavascript(script, null) }
        } else {
            pendingUpdates.add(script)
        }
    }

    // Menyesuaikan zoom agar seluruh anggota terlihat
    fun fitAll() {
        val script = "fitAllMembers();"
        if (isMapLoaded) {
            post { webView.evaluateJavascript(script, null) }
        } else {
            pendingUpdates.add(script)
        }
    }
}
