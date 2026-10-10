package com.fujiwaracreative.famloop.ui.parent

import android.content.Intent
import android.graphics.Color
import android.graphics.Typeface
import android.graphics.drawable.GradientDrawable
import android.os.Bundle
import android.view.Gravity
import android.view.View
import android.widget.*
import androidx.appcompat.app.AppCompatActivity
import com.fujiwaracreative.famloop.data.local.SessionManager
import com.fujiwaracreative.famloop.data.remote.CreateGroupRequest
import com.fujiwaracreative.famloop.data.remote.LoginRequest
import com.fujiwaracreative.famloop.data.remote.NetworkClient
import com.fujiwaracreative.famloop.data.remote.RegisterRequest
import com.fujiwaracreative.famloop.ui.MainActivity
import kotlinx.coroutines.CoroutineScope
import kotlinx.coroutines.Dispatchers
import kotlinx.coroutines.launch
import kotlinx.coroutines.withContext

class ParentAuthActivity : AppCompatActivity() {

    private lateinit var session: SessionManager
    private var isLoginMode = true

    // UI Elements
    private lateinit var tabLoginBtn: Button
    private lateinit var tabRegisterBtn: Button
    private lateinit var loginFormLayout: LinearLayout
    private lateinit var registerFormLayout: LinearLayout
    private lateinit var groupFormLayout: LinearLayout
    private lateinit var progressLoading: ProgressBar

    // Login inputs
    private lateinit var etLoginUsername: EditText
    private lateinit var etLoginPass: EditText

    // Register inputs
    private lateinit var etRegUsername: EditText
    private lateinit var etRegEmail: EditText
    private lateinit var etRegPhone: EditText
    private lateinit var etRegPass: EditText

    // Create group inputs
    private lateinit var etGroupName: EditText

    override fun onCreate(savedInstanceState: Bundle?) {
        super.onCreate(savedInstanceState)
        session = SessionManager.getInstance(this)

        buildAuthView()
    }

    private fun buildAuthView() {
        val scrollView = ScrollView(this).apply {
            setBackgroundColor(Color.parseColor("#F8FAFC"))
            isFillViewport = true
        }

        val container = LinearLayout(this).apply {
            orientation = LinearLayout.VERTICAL
            gravity = Gravity.CENTER_HORIZONTAL
            setPadding(48, 64, 48, 64)
        }

        // Header Title
        val title = TextView(this).apply {
            text = "Akun Orang Tua"
            textSize = 24f
            setTypeface(null, Typeface.BOLD)
            setTextColor(Color.parseColor("#0F172A"))
            gravity = Gravity.CENTER
        }
        container.addView(title)

        val subtitle = TextView(this).apply {
            text = "Masuk atau daftar untuk memantau lokasi keluarga"
            textSize = 14f
            setTextColor(Color.parseColor("#64748B"))
            gravity = Gravity.CENTER
            setPadding(0, 8, 0, 36)
        }
        container.addView(subtitle)

        // Tab Toggle Container
        val tabContainer = LinearLayout(this).apply {
            orientation = LinearLayout.HORIZONTAL
            val bg = GradientDrawable().apply {
                setColor(Color.parseColor("#E2E8F0"))
                cornerRadius = 16f
            }
            background = bg
            setPadding(6, 6, 6, 6)
            layoutParams = LinearLayout.LayoutParams(
                LinearLayout.LayoutParams.MATCH_PARENT,
                LinearLayout.LayoutParams.WRAP_CONTENT
            )
        }

        tabLoginBtn = Button(this).apply {
            text = "Masuk"
            textSize = 14f
            setTypeface(null, Typeface.BOLD)
            layoutParams = LinearLayout.LayoutParams(0, LinearLayout.LayoutParams.WRAP_CONTENT, 1f)
            setOnClickListener { switchTab(true) }
        }
        tabRegisterBtn = Button(this).apply {
            text = "Daftar Baru"
            textSize = 14f
            setTypeface(null, Typeface.BOLD)
            layoutParams = LinearLayout.LayoutParams(0, LinearLayout.LayoutParams.WRAP_CONTENT, 1f)
            setOnClickListener { switchTab(false) }
        }

        tabContainer.addView(tabLoginBtn)
        tabContainer.addView(tabRegisterBtn)
        container.addView(tabContainer)

        // Loading spinner
        progressLoading = ProgressBar(this).apply {
            visibility = View.GONE
            setPadding(0, 24, 0, 24)
        }
        container.addView(progressLoading)

        // Form Login
        loginFormLayout = buildLoginForm()
        container.addView(loginFormLayout)

        // Form Register
        registerFormLayout = buildRegisterForm()
        container.addView(registerFormLayout)

        // Form Create Group (Muncul bila user baru belum punya grup)
        groupFormLayout = buildGroupForm()
        groupFormLayout.visibility = View.GONE
        container.addView(groupFormLayout)

        // Back to Main Button
        val backBtn = Button(this).apply {
            text = "← Kembali ke Pemilih Peran"
            setTextColor(Color.parseColor("#64748B"))
            setBackgroundColor(Color.TRANSPARENT)
            textSize = 13f
            setPadding(0, 48, 0, 0)
            setOnClickListener {
                val intent = Intent(this@ParentAuthActivity, MainActivity::class.java).apply {
                    putExtra("EXTRA_FORCE_CHOOSER", true)
                }
                startActivity(intent)
                finish()
            }
        }
        container.addView(backBtn)

        scrollView.addView(container)
        setContentView(scrollView)

        switchTab(true)
    }

    private fun switchTab(login: Boolean) {
        isLoginMode = login
        if (login) {
            tabLoginBtn.setBackgroundColor(Color.WHITE)
            tabLoginBtn.setTextColor(Color.parseColor("#0F172A"))
            tabRegisterBtn.setBackgroundColor(Color.TRANSPARENT)
            tabRegisterBtn.setTextColor(Color.parseColor("#64748B"))

            loginFormLayout.visibility = View.VISIBLE
            registerFormLayout.visibility = View.GONE
            groupFormLayout.visibility = View.GONE
        } else {
            tabRegisterBtn.setBackgroundColor(Color.WHITE)
            tabRegisterBtn.setTextColor(Color.parseColor("#0F172A"))
            tabLoginBtn.setBackgroundColor(Color.TRANSPARENT)
            tabLoginBtn.setTextColor(Color.parseColor("#64748B"))

            loginFormLayout.visibility = View.GONE
            registerFormLayout.visibility = View.VISIBLE
            groupFormLayout.visibility = View.GONE
        }
    }

    private fun buildLoginForm(): LinearLayout {
        return LinearLayout(this).apply {
            orientation = LinearLayout.VERTICAL
            setPadding(0, 36, 0, 0)
            layoutParams = LinearLayout.LayoutParams(
                LinearLayout.LayoutParams.MATCH_PARENT,
                LinearLayout.LayoutParams.WRAP_CONTENT
            )

            etLoginUsername = createInput("Username atau Email")
            addView(etLoginUsername)

            etLoginPass = createInput("Password", isPassword = true)
            addView(etLoginPass)

            val btnSubmit = createSubmitButton("Masuk") {
                handleLogin()
            }
            addView(btnSubmit)
        }
    }

    private fun buildRegisterForm(): LinearLayout {
        return LinearLayout(this).apply {
            orientation = LinearLayout.VERTICAL
            setPadding(0, 36, 0, 0)
            layoutParams = LinearLayout.LayoutParams(
                LinearLayout.LayoutParams.MATCH_PARENT,
                LinearLayout.LayoutParams.WRAP_CONTENT
            )

            etRegUsername = createInput("Username (mis: bapak_budi)")
            addView(etRegUsername)

            etRegEmail = createInput("Email (mis: budi@famloop.id)")
            addView(etRegEmail)

            etRegPhone = createInput("Nomor HP (mis: 081234567890)")
            addView(etRegPhone)

            etRegPass = createInput("Password (min 6 karakter)", isPassword = true)
            addView(etRegPass)

            val btnSubmit = createSubmitButton("Daftar Sekarang") {
                handleRegister()
            }
            addView(btnSubmit)
        }
    }

    private fun buildGroupForm(): LinearLayout {
        return LinearLayout(this).apply {
            orientation = LinearLayout.VERTICAL
            setPadding(0, 36, 0, 0)
            layoutParams = LinearLayout.LayoutParams(
                LinearLayout.LayoutParams.MATCH_PARENT,
                LinearLayout.LayoutParams.WRAP_CONTENT
            )

            val titleG = TextView(this@ParentAuthActivity).apply {
                text = "Buat Grup Keluarga"
                textSize = 18f
                setTypeface(null, Typeface.BOLD)
                setTextColor(Color.parseColor("#0F172A"))
                gravity = Gravity.CENTER
            }
            addView(titleG)

            val descG = TextView(this@ParentAuthActivity).apply {
                text = "Akun belum memiliki grup keluarga. Masukkan nama grup Anda:"
                textSize = 13f
                setTextColor(Color.parseColor("#64748B"))
                gravity = Gravity.CENTER
                setPadding(0, 8, 0, 24)
            }
            addView(descG)

            etGroupName = createInput("Nama Grup (misal: Keluarga Bahagia)")
            addView(etGroupName)

            val btnCreate = createSubmitButton("Simpan & Buka Peta") {
                handleCreateGroup()
            }
            addView(btnCreate)
        }
    }

    private fun createInput(hintText: String, isPassword: Boolean = false): EditText {
        return EditText(this).apply {
            hint = hintText
            textSize = 14f
            setPadding(32, 28, 32, 28)
            val bg = GradientDrawable().apply {
                setColor(Color.WHITE)
                cornerRadius = 16f
                setStroke(2, Color.parseColor("#CBD5E1"))
            }
            background = bg
            if (isPassword) {
                inputType = android.text.InputType.TYPE_CLASS_TEXT or android.text.InputType.TYPE_TEXT_VARIATION_PASSWORD
            }
            layoutParams = LinearLayout.LayoutParams(
                LinearLayout.LayoutParams.MATCH_PARENT,
                LinearLayout.LayoutParams.WRAP_CONTENT
            ).apply {
                bottomMargin = 24
            }
        }
    }

    private fun createSubmitButton(label: String, onClick: () -> Unit): Button {
        return Button(this).apply {
            text = label
            setTextColor(Color.WHITE)
            textSize = 15f
            setTypeface(null, Typeface.BOLD)
            val bg = GradientDrawable().apply {
                setColor(Color.parseColor("#2563EB"))
                cornerRadius = 16f
            }
            background = bg
            setPadding(32, 28, 32, 28)
            setOnClickListener { onClick() }
            layoutParams = LinearLayout.LayoutParams(
                LinearLayout.LayoutParams.MATCH_PARENT,
                LinearLayout.LayoutParams.WRAP_CONTENT
            ).apply {
                topMargin = 16
            }
        }
    }

    private fun handleLogin() {
        val userOrEmail = etLoginUsername.text.toString().trim()
        val pass = etLoginPass.text.toString().trim()

        if (userOrEmail.isEmpty() || pass.isEmpty()) {
            Toast.makeText(this, "Silakan isi username/email dan password", Toast.LENGTH_SHORT).show()
            return
        }

        setLoading(true)
        CoroutineScope(Dispatchers.IO).launch {
            try {
                val resp = NetworkClient.apiService.login(LoginRequest(userOrEmail, pass))
                withContext(Dispatchers.Main) {
                    setLoading(false)
                    if (resp.isSuccessful && resp.body()?.data != null) {
                        val authData = resp.body()!!.data!!
                        saveSessionAndProceed(authData.token, authData.user.id, authData.user.username)
                    } else {
                        val errMsg = resp.errorBody()?.string() ?: "Login gagal"
                        Toast.makeText(this@ParentAuthActivity, "Gagal: $errMsg", Toast.LENGTH_LONG).show()
                    }
                }
            } catch (e: Exception) {
                withContext(Dispatchers.Main) {
                    setLoading(false)
                    Toast.makeText(this@ParentAuthActivity, "Error koneksi: ${e.message}", Toast.LENGTH_LONG).show()
                }
            }
        }
    }

    private fun handleRegister() {
        val user = etRegUsername.text.toString().trim()
        val email = etRegEmail.text.toString().trim()
        val phone = etRegPhone.text.toString().trim()
        val pass = etRegPass.text.toString().trim()

        if (user.isEmpty() || email.isEmpty() || phone.isEmpty() || pass.isEmpty()) {
            Toast.makeText(this, "Semua field wajib diisi", Toast.LENGTH_SHORT).show()
            return
        }

        setLoading(true)
        CoroutineScope(Dispatchers.IO).launch {
            try {
                val resp = NetworkClient.apiService.register(RegisterRequest(user, email, phone, pass))
                withContext(Dispatchers.Main) {
                    setLoading(false)
                    if (resp.isSuccessful && resp.body()?.data != null) {
                        val authData = resp.body()!!.data!!
                        Toast.makeText(this@ParentAuthActivity, "Pendaftaran berhasil!", Toast.LENGTH_SHORT).show()
                        saveSessionAndProceed(authData.token, authData.user.id, authData.user.username)
                    } else {
                        val errMsg = resp.errorBody()?.string() ?: "Pendaftaran gagal"
                        Toast.makeText(this@ParentAuthActivity, "Gagal: $errMsg", Toast.LENGTH_LONG).show()
                    }
                }
            } catch (e: Exception) {
                withContext(Dispatchers.Main) {
                    setLoading(false)
                    Toast.makeText(this@ParentAuthActivity, "Error koneksi: ${e.message}", Toast.LENGTH_LONG).show()
                }
            }
        }
    }

    private fun saveSessionAndProceed(token: String, userId: String, username: String) {
        session.token = token
        session.userId = userId
        session.username = username
        session.role = "parent"
        NetworkClient.authToken = token

        // Periksa apakah user sudah memiliki grup
        checkExistingGroups()
    }

    private fun checkExistingGroups() {
        setLoading(true)
        CoroutineScope(Dispatchers.IO).launch {
            try {
                val resp = NetworkClient.apiService.getGroups()
                withContext(Dispatchers.Main) {
                    setLoading(false)
                    if (resp.isSuccessful) {
                        val groups = resp.body()?.data
                        if (!groups.isNullOrEmpty()) {
                            val activeGroup = groups[0]
                            session.groupId = activeGroup.id
                            session.groupName = activeGroup.name
                            goToLiveMap(activeGroup.id, activeGroup.name)
                        } else {
                            // Tampilkan form buat grup
                            showCreateGroupForm()
                        }
                    } else {
                        showCreateGroupForm()
                    }
                }
            } catch (_: Exception) {
                withContext(Dispatchers.Main) {
                    setLoading(false)
                    showCreateGroupForm()
                }
            }
        }
    }

    private fun showCreateGroupForm() {
        loginFormLayout.visibility = View.GONE
        registerFormLayout.visibility = View.GONE
        groupFormLayout.visibility = View.VISIBLE
    }

    private fun handleCreateGroup() {
        val groupName = etGroupName.text.toString().trim()
        if (groupName.isEmpty()) {
            Toast.makeText(this, "Nama grup wajib diisi", Toast.LENGTH_SHORT).show()
            return
        }

        setLoading(true)
        CoroutineScope(Dispatchers.IO).launch {
            try {
                val resp = NetworkClient.apiService.createGroup(CreateGroupRequest(groupName))
                withContext(Dispatchers.Main) {
                    setLoading(false)
                    if (resp.isSuccessful && resp.body()?.data != null) {
                        val newGroup = resp.body()!!.data!!
                        session.groupId = newGroup.id
                        session.groupName = newGroup.name
                        Toast.makeText(this@ParentAuthActivity, "Grup '${newGroup.name}' berhasil dibuat!", Toast.LENGTH_SHORT).show()
                        goToLiveMap(newGroup.id, newGroup.name)
                    } else {
                        val errMsg = resp.errorBody()?.string() ?: "Gagal membuat grup"
                        Toast.makeText(this@ParentAuthActivity, errMsg, Toast.LENGTH_LONG).show()
                    }
                }
            } catch (e: Exception) {
                withContext(Dispatchers.Main) {
                    setLoading(false)
                    Toast.makeText(this@ParentAuthActivity, "Error koneksi: ${e.message}", Toast.LENGTH_LONG).show()
                }
            }
        }
    }

    private fun goToLiveMap(groupId: String, groupName: String) {
        val intent = Intent(this, LiveMapActivity::class.java).apply {
            putExtra("EXTRA_GROUP_ID", groupId)
            putExtra("EXTRA_GROUP_NAME", groupName)
        }
        startActivity(intent)
        finish()
    }

    private fun setLoading(loading: Boolean) {
        progressLoading.visibility = if (loading) View.VISIBLE else View.GONE
    }
}
