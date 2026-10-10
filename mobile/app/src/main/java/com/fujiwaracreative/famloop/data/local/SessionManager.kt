package com.fujiwaracreative.famloop.data.local

import android.content.Context
import android.content.SharedPreferences

class SessionManager(context: Context) {

    private val prefs: SharedPreferences =
        context.getSharedPreferences(PREF_NAME, Context.MODE_PRIVATE)

    companion object {
        private const val PREF_NAME = "famloop_session"
        private const val KEY_TOKEN = "jwt_token"
        private const val KEY_USER_ID = "user_id"
        private const val KEY_USERNAME = "username"
        private const val KEY_GROUP_ID = "group_id"
        private const val KEY_GROUP_NAME = "group_name"
        private const val KEY_ROLE = "user_role" // "parent" or "member"

        @Volatile
        private var instance: SessionManager? = null

        fun getInstance(context: Context): SessionManager =
            instance ?: synchronized(this) {
                instance ?: SessionManager(context.applicationContext).also { instance = it }
            }
    }

    var token: String?
        get() = prefs.getString(KEY_TOKEN, null)
        set(value) = prefs.edit().putString(KEY_TOKEN, value).apply()

    var userId: String?
        get() = prefs.getString(KEY_USER_ID, null)
        set(value) = prefs.edit().putString(KEY_USER_ID, value).apply()

    var username: String?
        get() = prefs.getString(KEY_USERNAME, null)
        set(value) = prefs.edit().putString(KEY_USERNAME, value).apply()

    var groupId: String?
        get() = prefs.getString(KEY_GROUP_ID, null)
        set(value) = prefs.edit().putString(KEY_GROUP_ID, value).apply()

    var groupName: String?
        get() = prefs.getString(KEY_GROUP_NAME, null)
        set(value) = prefs.edit().putString(KEY_GROUP_NAME, value).apply()

    var role: String?
        get() = prefs.getString(KEY_ROLE, null)
        set(value) = prefs.edit().putString(KEY_ROLE, value).apply()

    val isLoggedIn: Boolean
        get() = !token.isNullOrEmpty() && !userId.isNullOrEmpty()

    fun clearSession() {
        prefs.edit().clear().apply()
    }
}
