# Phase 1: Go Backend Infrastructure, Domain Layer & SQLite Storage

## Fokus Phase 1
Membangun fondasi backend dengan Clean Architecture, model domain murni, koneksi SQLite WAL mode (pure Go tanpa CGO), serta implementasi repository database.

---

## Daftar Task & Langkah Eksekusi

### Task 1.1: Pembaruan Go Module & Dependensi
- **File**: `go.mod`
- **Tujuan**: Memasang library esensial:
  - `modernc.org/sqlite` (Pure Go driver SQLite, aman tanpa GCC/CGO di Windows/Linux).
  - `golang.org/x/crypto/bcrypt` (Hashing password).
  - `github.com/golang-jwt/jwt/v5` (Token otentikasi orang tua).
  - `github.com/google/uuid` (UUID generator).
- **Verifikasi**: Jalankan `go mod tidy` dan pastikan compile lulus.

---

### Task 1.2: Pembuatan Entitas Domain Murni (Inward Core)
- **File Baru**:
  - `internal/domain/user.go`: Entitas User orang tua (`ID`, `Username`, `Email`, `PhoneNumber`, `PasswordHash`, `CreatedAt`).
  - `internal/domain/group.go`: Entitas FamilyGroup (`ID`, `Name`, `ParentID`, `InviteCode`, `CreatedAt`).
  - `internal/domain/member_device.go`: Entitas MemberDevice (`ID`, `GroupID`, `DeviceUUID`, `DeviceName`, `IPv4`, `IPv6`, `MACAddress`, `BatteryLevel`, `IsActive`, `LastSeenAt`).
  - `internal/domain/location.go`: Entitas LocationPoint (`ID`, `DeviceID`, `GroupID`, `Latitude`, `Longitude`, `Accuracy`, `Speed`, `Heading`, `RecordedAt`).
- **File Interface Repository**:
  - `internal/domain/repositories.go`: Interface kontrak:
    - `UserRepository`: `Create`, `GetByID`, `GetByEmail`, `GetByUsername`.
    - `GroupRepository`: `Create`, `GetByID`, `GetByParentID`, `AddMemberDevice`, `GetMemberDevices`, `FindDeviceInGroup`.
    - `LocationRepository`: `SavePoint`, `SaveBatch`, `GetLatestGroupLocations`, `GetDeviceHistory`.

---

### Task 1.3: Inisialisasi Database SQLite Adapter
- **File Baru**: `internal/repository/sqlite/db.go`
- **Konfigurasi Engine**:
  - Buka koneksi file `database/famloop.db`.
  - Aktifkan mode WAL: `PRAGMA journal_mode = WAL;`.
  - Pasang timeout: `PRAGMA busy_timeout = 5000;`.
  - Aktifkan foreign key: `PRAGMA foreign_keys = ON;`.
  - Jalankan script DDL auto-migration untuk tabel `users`, `family_groups`, `member_devices`, dan `location_history` beserta index uniknya.

---

### Task 1.4: Implementasi Repository SQLite
- **File Baru**:
  - `internal/repository/sqlite/user_repository.go`
  - `internal/repository/sqlite/group_repository.go`
  - `internal/repository/sqlite/location_repository.go`
- **Karakteristik Kritis**:
  - `SaveBatch`: Mendukung query idempotency `INSERT OR IGNORE` berdasarkan `(device_id, recorded_at)` agar sync offline dari Android tidak menimbulkan duplikasi saat dicoba ulang berkali-kali.

---

## Verifikasi Phase 1
- Buat unit test SQLite in-memory:
  `go test -v ./internal/repository/sqlite/...`
- Pastikan tabel terbuat, relasi foreign key berfungsi, dan batch insert berjalan cepat.
