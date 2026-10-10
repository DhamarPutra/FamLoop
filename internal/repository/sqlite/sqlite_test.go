package sqlite

import (
	"context"
	"strings"
	"testing"
	"time"

	"famloop/internal/domain"
)

func TestUserRepository(t *testing.T) {
	ctx := context.Background()
	db, err := NewDB(":memory:")
	if err != nil {
		t.Fatalf("failed to init in-memory db: %v", err)
	}
	defer db.Close()

	userRepo := NewUserRepository(db)

	// 1. Create user without avatar -> should get random default avatar
	u := &domain.User{
		ID:           "user-1",
		Username:     "ayah_budi",
		Email:        "ayah@famloop.id",
		PhoneNumber:  "08123456789",
		PasswordHash: "hashed_secret",
	}

	if err := userRepo.Create(ctx, u); err != nil {
		t.Fatalf("create user failed: %v", err)
	}

	byID, err := userRepo.GetByID(ctx, "user-1")
	if err != nil || byID == nil {
		t.Fatalf("get by id failed: %v", err)
	}
	if !strings.HasPrefix(byID.Avatar, "char_") {
		t.Errorf("expected default random avatar starting with char_, got %s", byID.Avatar)
	}

	// 2. Upload custom PP
	if err := userRepo.UpdateAvatar(ctx, "user-1", "/uploads/avatars/ayah_photo.jpg"); err != nil {
		t.Fatalf("update avatar failed: %v", err)
	}

	updated, _ := userRepo.GetByID(ctx, "user-1")
	if updated.Avatar != "/uploads/avatars/ayah_photo.jpg" {
		t.Errorf("expected custom avatar /uploads/avatars/ayah_photo.jpg, got %s", updated.Avatar)
	}
}

func TestGroupAndMemberDeviceRepository(t *testing.T) {
	ctx := context.Background()
	db, err := NewDB(":memory:")
	if err != nil {
		t.Fatalf("failed to init in-memory db: %v", err)
	}
	defer db.Close()

	userRepo := NewUserRepository(db)
	groupRepo := NewGroupRepository(db)

	_ = userRepo.Create(ctx, &domain.User{
		ID:           "parent-1",
		Username:     "ibu_ani",
		Email:        "ibu@famloop.id",
		PhoneNumber:  "08199988877",
		PasswordHash: "hashed_secret",
	})

	g := &domain.FamilyGroup{
		ID:         "group-1",
		Name:       "Keluarga Bahagia",
		ParentID:   "parent-1",
		InviteCode: "FAM-9999",
	}
	if err := groupRepo.Create(ctx, g); err != nil {
		t.Fatalf("create group failed: %v", err)
	}

	// 1. Add device via QR scan data with empty avatar -> should get random default
	dev := &domain.MemberDevice{
		ID:           "dev-1",
		GroupID:      "group-1",
		DeviceUUID:   "android-hw-1234",
		DeviceName:   "Samsung A54 Anak",
		IPv4:         "192.168.1.50",
		IPv6:         "fe80::1",
		MACAddress:   "00:11:22:33:44:55",
		BatteryLevel: 85,
	}
	if err := groupRepo.AddMemberDevice(ctx, dev); err != nil {
		t.Fatalf("add member device failed: %v", err)
	}

	devices, err := groupRepo.GetMemberDevices(ctx, "group-1")
	if err != nil {
		t.Fatalf("get member devices failed: %v", err)
	}
	if len(devices) != 1 {
		t.Fatalf("expected 1 member device, got %d", len(devices))
	}
	if !strings.HasPrefix(devices[0].Avatar, "char_") {
		t.Errorf("expected default avatar starting with char_, got %s", devices[0].Avatar)
	}

	// 2. Child uploads custom avatar PP
	if err := groupRepo.UpdateDeviceAvatar(ctx, "dev-1", "/uploads/avatars/child_selfie.png"); err != nil {
		t.Fatalf("update device avatar failed: %v", err)
	}

	devAfter, _ := groupRepo.GetDeviceByID(ctx, "dev-1")
	if devAfter.Avatar != "/uploads/avatars/child_selfie.png" {
		t.Errorf("expected updated avatar /uploads/avatars/child_selfie.png, got %s", devAfter.Avatar)
	}
}

func TestLocationBatchDeduplication(t *testing.T) {
	ctx := context.Background()
	db, err := NewDB(":memory:")
	if err != nil {
		t.Fatalf("failed to init in-memory db: %v", err)
	}
	defer db.Close()

	userRepo := NewUserRepository(db)
	groupRepo := NewGroupRepository(db)
	locRepo := NewLocationRepository(db)

	_ = userRepo.Create(ctx, &domain.User{
		ID: "p-1", Username: "parent", Email: "p@fam.id", PhoneNumber: "081", PasswordHash: "x",
	})
	_ = groupRepo.Create(ctx, &domain.FamilyGroup{
		ID: "g-1", Name: "Fam", ParentID: "p-1",
	})
	_ = groupRepo.AddMemberDevice(ctx, &domain.MemberDevice{
		ID: "d-1", GroupID: "g-1", DeviceUUID: "uuid-1", DeviceName: "Anak 1",
	})

	nowMs := time.Now().UnixMilli()
	points := []domain.LocationPoint{
		{DeviceID: "d-1", GroupID: "g-1", Latitude: -6.200, Longitude: 106.800, RecordedAt: nowMs - 20000},
		{DeviceID: "d-1", GroupID: "g-1", Latitude: -6.201, Longitude: 106.801, RecordedAt: nowMs - 10000},
		{DeviceID: "d-1", GroupID: "g-1", Latitude: -6.202, Longitude: 106.802, RecordedAt: nowMs},
	}

	// 1. First batch save
	inserted, err := locRepo.SaveBatch(ctx, points)
	if err != nil {
		t.Fatalf("save batch failed: %v", err)
	}
	if inserted != 3 {
		t.Errorf("expected 3 inserted, got %d", inserted)
	}

	// 2. Retry sync with 2 duplicate points and 1 new point
	retryPoints := []domain.LocationPoint{
		{DeviceID: "d-1", GroupID: "g-1", Latitude: -6.201, Longitude: 106.801, RecordedAt: nowMs - 10000}, // Duplicate
		{DeviceID: "d-1", GroupID: "g-1", Latitude: -6.202, Longitude: 106.802, RecordedAt: nowMs},         // Duplicate
		{DeviceID: "d-1", GroupID: "g-1", Latitude: -6.203, Longitude: 106.803, RecordedAt: nowMs + 10000}, // New
	}

	insertedRetry, err := locRepo.SaveBatch(ctx, retryPoints)
	if err != nil {
		t.Fatalf("retry batch failed: %v", err)
	}
	if insertedRetry != 1 {
		t.Errorf("expected only 1 new point inserted, got %d", insertedRetry)
	}

	// 3. Test GetLatestGroupLocations includes Avatar
	latest, err := locRepo.GetLatestGroupLocations(ctx, "g-1")
	if err != nil {
		t.Fatalf("get latest locations failed: %v", err)
	}
	if len(latest) != 1 {
		t.Fatalf("expected 1 latest device location, got %d", len(latest))
	}
	if !strings.HasPrefix(latest[0].Avatar, "char_") {
		t.Errorf("expected latest avatar starting with char_, got %s", latest[0].Avatar)
	}
	if latest[0].Latitude != -6.203 {
		t.Errorf("expected latest latitude -6.203, got %f", latest[0].Latitude)
	}

	// 4. Test GetDeviceHistory
	history, err := locRepo.GetDeviceHistory(ctx, "d-1", 10)
	if err != nil {
		t.Fatalf("get history failed: %v", err)
	}
	if len(history) != 4 {
		t.Errorf("expected 4 total history points, got %d", len(history))
	}
}
