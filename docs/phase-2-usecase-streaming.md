# Phase 2: Application Use Cases & Real-Time Streaming Engine

## Fokus Phase 2
Membangun logika bisnis aplikasi (Use Cases) yang terisolasi dari framework web, pengelolaan sistem Avatar (random karakter default & upload custom PP), serta menyiapkan mesin streaming real-time Server-Sent Events (SSE) sebagai protokol utama yang tahan blokir firewall, didampingi embedded MQTT broker sebagai opsi transport ringan.

---

## Daftar Task & Langkah Eksekusi

### Task 2.1: Auth Use Case (Orang Tua / Kepala Keluarga)
- **File Baru**: `internal/usecase/auth_usecase.go`
- **Fitur**:
  - `Register(input RegisterInput) (*domain.User, error)`:
    - Validasi format email & nomor telepon.
    - Cek keunikan email, username, dan nomor telepon.
    - Tetapkan avatar random karakter default jika kosong (`domain.GetRandomDefaultAvatar()`).
    - Enkripsi password menggunakan `bcrypt.GenerateFromPassword(pass, 12)`.
  - `Login(usernameOrEmail, password string) (string, *domain.User, error)`:
    - Verifikasi hash password.
    - Terbitkan token JWT dengan masa aktif 30 hari.

---

### Task 2.2: Family Group & Device Pairing Use Case
- **File Baru**: `internal/usecase/group_usecase.go`
- **Fitur**:
  - `CreateGroup(parentID, groupName string) (*domain.FamilyGroup, error)`:
    - Generate kode unik grup (`invite_code`).
  - `GetParentGroups(parentID string) ([]domain.FamilyGroup, error)`.
  - `PairDeviceFromQR(parentID, groupID string, payload QRScanPayload) (*domain.MemberDevice, error)`:
    - Verifikasi bahwa `parentID` adalah pemilik `groupID`.
    - Simpan data identitas perangkat: `DeviceUUID`, `DeviceName`, `Avatar`, `IPv4`, `IPv6`, `MACAddress`.
    - Jika avatar kosong di payload, otomatis berikan avatar karakter acak.
    - Jika perangkat sudah terdaftar sebelumnya di grup tersebut, perbarui IP, avatar, dan status aktifnya (*upsert*).

---

### Task 2.3: Tracking & Offline Batch Sync Use Case
- **File Baru**:
  - `internal/usecase/tracking_usecase.go`:
    - `IngestLivePing(point domain.LocationPoint) error`:
      - Validasi batas latitude (-90 s/d 90) dan longitude (-180 s/d 180).
      - Simpan ke `location_history` dan perbarui `last_seen_at` di `member_devices`.
      - Kirim event ke SSE Hub agar peta orang tua dan anak langsung bergeser real-time.
  - `internal/usecase/sync_usecase.go`:
    - `SyncBatchOffline(groupID string, points []domain.LocationPoint) (int, error)`:
      - Validasi seluruh daftar koordinat.
      - Jalankan batch insert atomic ke SQLite.
      - Return total titik baru yang berhasil disimpan (mengabaikan duplikat).

---

### Task 2.4: Avatar Management Use Case (Upload Foto Profil)
- **File Baru**: `internal/usecase/avatar_usecase.go`
- **Fitur**:
  - `UploadUserAvatar(userID string, fileHeader io.Reader, filename string) (string, error)`:
    - Validasi format file (JPEG, PNG, WebP) dan batas ukuran (maksimal 5MB).
    - Simpan ke direktori `uploads/avatars/user_{id}_{timestamp}.ext`.
    - Perbarui kolom `avatar` di tabel `users`.
  - `UploadDeviceAvatar(deviceID, groupID string, fileHeader io.Reader, filename string) (string, error)`:
    - Simpan ke direktori `uploads/avatars/dev_{id}_{timestamp}.ext`.
    - Perbarui kolom `avatar` di tabel `member_devices`.
    - Kirim sinyal broadcast SSE event `MemberAvatarUpdated` ke grup keluarga agar seluruh peta langsung mengupdate marker avatar.

---

### Task 2.5: Server-Sent Events (SSE) Broadcast Hub
- **File Baru**: `internal/transport/sse/hub.go`
- **Mekanisme**:
  - Hub mengelola channel subscriber per `group_id`.
  - Menyiarkan 2 jenis event utama:
    - `LocationUpdated`: Koordinat dan status baterai anggota.
    - `MemberAvatarUpdated`: Perubahan foto profil / avatar anggota.
  - Ticker keep-alive otomatis mengirim komentar `: ping` tiap 15 detik agar koneksi HTTP tidak ditutup oleh timeout ISP atau gateway seluler.

---

### Task 2.6: Embedded Mochi-MQTT Broker (Opsional / Sekunder)
- **File Baru**: `internal/transport/mqtt/broker.go`
- **Mekanisme**:
  - Inisialisasi `mochi-mqtt/server/v2` pada port `:1883`.
  - Pasang hook listener pada topic `famloop/devices/+/location`.
  - Teruskan paket lokasi yang diterima via MQTT ke `tracking_usecase` dan siarkan ke SSE Hub.

---

## Verifikasi Phase 2
- Unit test Use Case menggunakan mock repository:
  `go test -v ./internal/usecase/...`
- Verifikasi upload foto profil, hashing, token signing, validasi koordinat, dan filter kepemilikan grup.
