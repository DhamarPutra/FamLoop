# Phase 4: Android Kotlin Native Client Architecture & Leaflet Maps

## Fokus Phase 4
Membangun aplikasi Android Kotlin Native: background foreground service hemat baterai, penyimpanan lokal offline-first menggunakan Room SQLite, sinkronisasi otomatis via WorkManager, generator QR anggota, pemindai QR berbasis CameraX untuk orang tua, serta **Peta Interaktif Real-Time Leaflet.js dengan Custom Avatar Bubbles untuk seluruh anggota keluarga (Orang Tua dan Anak)**.

---

## Daftar Task & Langkah Eksekusi

### Task 4.1: Android Manifest & Konfigurasi Izin
- **File**: `mobile/app/src/main/AndroidManifest.xml`
- **Izin Esensial**:
  ```xml
  <uses-permission android:name="android.permission.INTERNET" />
  <uses-permission android:name="android.permission.ACCESS_NETWORK_STATE" />
  <uses-permission android:name="android.permission.ACCESS_FINE_LOCATION" />
  <uses-permission android:name="android.permission.ACCESS_COARSE_LOCATION" />
  <uses-permission android:name="android.permission.ACCESS_BACKGROUND_LOCATION" />
  <uses-permission android:name="android.permission.FOREGROUND_SERVICE" />
  <uses-permission android:name="android.permission.FOREGROUND_SERVICE_LOCATION" />
  <uses-permission android:name="android.permission.POST_NOTIFICATIONS" />
  <uses-permission android:name="android.permission.CAMERA" />
  ```

---

### Task 4.2: Room SQLite Offline Buffer
- **File**:
  - `mobile/app/src/main/java/com/famloop/data/local/LocationEntity.kt`:
    Entity tabel `offline_locations` dengan atribut: `id`, `latitude`, `longitude`, `accuracy`, `speed`, `heading`, `recordedAt`, `isSynced`.
  - `mobile/app/src/main/java/com/famloop/data/local/LocationDao.kt`:
    Method `insertPoint`, `getUnsyncedPoints(limit)`, `markAsSynced(ids)`, `deleteOldSynced()`.
  - `mobile/app/src/main/java/com/famloop/data/local/AppDatabase.kt`:
    Database Room lokal di perangkat Android.

---

### Task 4.3: Background Tracking Foreground Service
- **File**: `mobile/app/src/main/java/com/famloop/service/LocationForegroundService.kt`
- **Mekanisme**:
  - Berjalan sebagai Android Foreground Service dengan notifikasi persisten (`NotificationCompat.Builder`).
  - Menggunakan `FusedLocationProviderClient` dengan interval adaptif:
    - Interval aktif bergerak: tiap 10-15 detik.
    - Interval diam: diperlambat hingga 60 detik untuk efisiensi baterai.
  - Setiap titik koordinat langsung disimpan ke Room SQLite lokal (`isSynced = false`).
  - Jika internet tersedia: coba kirim via REST API. Jika gagal: biarkan di Room.

---

### Task 4.4: Auto-Sync Worker (Offline-First Recovery)
- **File**: `mobile/app/src/main/java/com/famloop/worker/LocationSyncWorker.kt`
- **Mekanisme**:
  - Didaftarkan ke `WorkManager` dengan batasan `Constraints.Builder().setRequiredNetworkType(NetworkType.CONNECTED).build()`.
  - Saat perangkat kembali mendapatkan koneksi internet, worker otomatis membaca seluruh titik `isSynced = false` dan mengirimkan payload ke `POST /api/v1/tracking/sync/batch`.
  - Setelah server membalas status 200 OK, titik-titik tersebut ditandai `isSynced = true`.

---

### Task 4.5: QR Generator & CameraX Scanner Flow
- **Hp Anggota / Anak (Tampilkan QR)**:
  - `mobile/app/src/main/java/com/famloop/ui/member/MemberQrActivity.kt`
  - Mengambil Hardware Device ID (`Settings.Secure.ANDROID_ID`), nama model ponsel, avatar pilihan, dan IP v4/v6 lokal.
  - Membentuk string JSON dan me-render QR Code di layar menggunakan ZXing Android Embedded.
- **Hp Orang Tua (Scan QR & Tambah Anggota)**:
  - `mobile/app/src/main/java/com/famloop/ui/parent/QrScannerActivity.kt`
  - Membuka kamera dengan CameraX dan ML Kit Barcode Scanning.
  - Membaca payload QR lalu mengirim ke API backend `POST /api/v1/groups/{group_id}/pair-device`.

---

### Task 4.6: Leaflet.js WebView Map dengan Custom Avatar Marker (Untuk Orang Tua & Anak)
- **Asset Web Lokal**: `mobile/app/src/main/assets/map/index.html`
  - Menggunakan library **Leaflet.js** offline-bundled + OpenStreetMap tile layer (100% gratis, tanpa API Key Google Maps).
  - CSS Marker Avatar Kustom:
    ```css
    .avatar-marker {
        width: 48px;
        height: 48px;
        border-radius: 50%;
        border: 3px solid #2563EB;
        box-shadow: 0 4px 10px rgba(0,0,0,0.3);
        background-size: cover;
        background-position: center;
        position: relative;
    }
    .battery-badge {
        position: absolute;
        bottom: -2px;
        right: -2px;
        background: #10B981;
        color: white;
        font-size: 10px;
        padding: 2px 4px;
        border-radius: 6px;
    }
    ```
- **Komponen Kotlin Reusable**: `mobile/app/src/main/java/com/famloop/ui/common/LeafletMapView.kt`
  - Menghubungkan Android WebView dengan Leaflet via Javascript Interface:
    - `updateMemberMarker(deviceId, name, lat, lng, avatarUrl, battery, isOnline)`: Menggeser avatar marker secara halus dengan animasi interpolasi pan.
    - `centerToMember(lat, lng)`.
  - Koneksi SSE real-time (`OkHttp EventSource`) dipasang di **ParentMapActivity** dan **MemberMapActivity**, sehingga **baik orang tua maupun anak bisa saling melihat posisi anggota keluarga lainnya di peta Leaflet secara bersamaan**.

---

## Verifikasi Phase 4
- Uji simpan titik lokasi pada mode Airplane (offline).
- Nyalakan kembali koneksi internet, verifikasi data terunggah secara utuh tanpa ada titik yang hilang ke backend.
- Uji render Leaflet WebView di hp orang tua dan hp anak, pastikan avatar tiap anggota muncul dan bergeser saat ada event SSE lokasi baru.
