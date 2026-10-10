# Phase 3: REST API Handlers, Streaming HTTP & Server Wiring

## Fokus Phase 3
Menyediakan HTTP presentation layer menggunakan Go Fiber v2, middleware otentikasi JWT, endpoint registrasi & pairing scan QR, endpoint upload foto profil avatar, endpoint sinkronisasi batch offline, penyajian static file gambar avatar, serta endpoint Server-Sent Events (SSE).

---

## Daftar Task & Langkah Eksekusi

### Task 3.1: Middleware & Standarisasi Error
- **File Baru**: `internal/transport/http/middleware.go`
- **Fitur**:
  - `JWTMiddleware`: Memvalidasi header `Authorization: Bearer <token>`, mengekstrak `user_id` ke context request.
  - `CORS`: Izinkan akses cross-origin dengan header lengkap.
  - `ErrorHandler`: Respons error terstandarisasi:
    ```json
    {
      "error": {
        "code": 400,
        "message": "Deskripsi kesalahan"
      }
    }
    ```

---

### Task 3.2: Auth & Family Group Handlers
- **File Baru**:
  - `internal/transport/http/auth_handler.go`:
    - `POST /api/v1/auth/register`
    - `POST /api/v1/auth/login`
    - `GET /api/v1/auth/me`
    - `POST /api/v1/auth/avatar` (Upload foto profil orang tua, multipart form data)
  - `internal/transport/http/group_handler.go`:
    - `POST /api/v1/groups` (Buat grup baru)
    - `GET /api/v1/groups` (Daftar grup milik orang tua)
    - `GET /api/v1/groups/:id` (Detail grup beserta daftar perangkat anggota)
    - `POST /api/v1/groups/:id/pair-device` (Pairing perangkat hasil scan QR hp anak)
    - `POST /api/v1/devices/:id/avatar` (Upload foto profil perangkat anak)

---

### Task 3.3: Tracking & Offline Batch Sync Handlers
- **File Baru**: `internal/transport/http/tracking_handler.go`
- **Fitur**:
  - `POST /api/v1/tracking/ping`: Menerima 1 titik koordinat aktif dari hp anggota.
  - `POST /api/v1/tracking/sync/batch`: Menerima array titik lokasi offline dari hp anggota yang baru tersambung internet.
  - `GET /api/v1/groups/:id/locations/latest`: Mengambil koordinat terakhir seluruh anggota grup beserta data Avatar untuk render peta Leaflet (dapat diakses oleh hp orang tua dan hp anak).
  - `GET /api/v1/groups/:id/members/:device_id/history`: Mengambil jejak perjalanan anggota.
  - `GET /api/v1/groups/:id/sse`: Endpoint Server-Sent Events untuk streaming lokasi dan perubahan avatar ke seluruh hp keluarga (orang tua & anak) secara real-time.

---

### Task 3.4: Router Setup, Static Serving & Dependency Injection Server
- **File Baru**:
  - `internal/transport/http/router.go`:
    - Registrasi seluruh route ke Fiber app.
    - Pasang static file handler: `app.Static("/uploads", "./uploads")` untuk akses publik file avatar yang diupload.
  - `cmd/server/main.go`:
    - Membaca environment variables (`PORT`, `DB_PATH`, `JWT_SECRET`, `UPLOAD_DIR`).
    - Membuat folder `./uploads/avatars` jika belum ada.
    - Menginisialisasi Database SQLite WAL.
    - Menginisialisasi Repositories -> Use Cases -> Handlers.
    - Menjalankan background MQTT Broker & SSE Hub.
    - Mendengarkan graceful shutdown OS signal.

---

## Verifikasi Phase 3
- Integration test API via Go HTTP client:
  `go test -v ./internal/transport/http/...`
- Pengujian upload file avatar multipart dan download file statisnya.
