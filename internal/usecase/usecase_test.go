package usecase

import (
	"context"
	"strings"
	"testing"
	"time"

	"famloop/internal/repository/sqlite"
	"famloop/internal/transport/sse"
)

func setupTestEnv(t *testing.T) (*AuthUseCase, *GroupUseCase, *TrackingUseCase, *SyncUseCase, *AvatarUseCase, *sse.Hub) {
	db, err := sqlite.NewDB(":memory:")
	if err != nil {
		t.Fatalf("failed to init in-memory db: %v", err)
	}

	userRepo := sqlite.NewUserRepository(db)
	groupRepo := sqlite.NewGroupRepository(db)
	locRepo := sqlite.NewLocationRepository(db)
	hub := sse.NewHub()

	jwtSecret := "super-secret-test-key-12345"
	authUC := NewAuthUseCase(userRepo, jwtSecret)
	groupUC := NewGroupUseCase(groupRepo, hub)
	trackingUC := NewTrackingUseCase(locRepo, groupRepo, hub)
	syncUC := NewSyncUseCase(locRepo, groupRepo, hub)
	avatarUC := NewAvatarUseCase(userRepo, groupRepo, hub)

	return authUC, groupUC, trackingUC, syncUC, avatarUC, hub
}

func TestAuthUseCase(t *testing.T) {
	ctx := context.Background()
	authUC, _, _, _, _, _ := setupTestEnv(t)

	// 1. Validation test
	_, err := authUC.Register(ctx, RegisterInput{
		Username: "ab", Email: "invalid-email", PhoneNumber: "123", Password: "123",
	})
	if err == nil {
		t.Fatal("expected validation error, got nil")
	}

	// 2. Successful registration with default random avatar
	regResult, err := authUC.Register(ctx, RegisterInput{
		Username:    "budi_santoso",
		Email:       "budi@famloop.id",
		PhoneNumber: "081234567890",
		Password:    "rahasia123",
	})
	if err != nil {
		t.Fatalf("register failed: %v", err)
	}
	if regResult.Token == "" {
		t.Fatal("expected non-empty token")
	}
	if !strings.HasPrefix(regResult.User.Avatar, "char_") {
		t.Errorf("expected default random avatar starting with char_, got %s", regResult.User.Avatar)
	}

	// 3. Duplicate check
	_, err = authUC.Register(ctx, RegisterInput{
		Username:    "budi_santoso",
		Email:       "budi2@famloop.id",
		PhoneNumber: "081999999999",
		Password:    "rahasia123",
	})
	if err != ErrUserAlreadyExists {
		t.Fatalf("expected ErrUserAlreadyExists, got %v", err)
	}

	// 4. Login with email
	loginResult, err := authUC.Login(ctx, LoginInput{
		UsernameOrEmail: "budi@famloop.id",
		Password:        "rahasia123",
	})
	if err != nil {
		t.Fatalf("login failed: %v", err)
	}
	if loginResult.User.Username != "budi_santoso" {
		t.Errorf("expected user budi_santoso, got %s", loginResult.User.Username)
	}

	// 5. Login invalid password
	_, err = authUC.Login(ctx, LoginInput{
		UsernameOrEmail: "budi_santoso",
		Password:        "wrong_password",
	})
	if err != ErrInvalidCredentials {
		t.Fatalf("expected ErrInvalidCredentials, got %v", err)
	}

	// 6. Token validation
	claims, err := authUC.ValidateToken(loginResult.Token)
	if err != nil {
		t.Fatalf("validate token failed: %v", err)
	}
	if claims.Username != "budi_santoso" {
		t.Errorf("expected username budi_santoso in claims, got %s", claims.Username)
	}
}

func TestGroupAndPairingUseCase(t *testing.T) {
	ctx := context.Background()
	authUC, groupUC, _, _, _, hub := setupTestEnv(t)

	authRes, err := authUC.Register(ctx, RegisterInput{
		Username: "ayah_cemara", Email: "ayah@fam.id", PhoneNumber: "081111111111", Password: "password123",
	})
	if err != nil || authRes == nil {
		t.Fatalf("register failed in test: %v", err)
	}
	parentID := authRes.User.ID

	// 1. Create group
	group, err := groupUC.CreateGroup(ctx, parentID, "Keluarga Cemara")
	if err != nil {
		t.Fatalf("create group failed: %v", err)
	}
	if !strings.HasPrefix(group.InviteCode, "FAM-") {
		t.Errorf("expected invite code starting with FAM-, got %s", group.InviteCode)
	}

	// 2. Subscribe to SSE events
	sseCh, unsubscribe := hub.Subscribe(group.ID)
	defer unsubscribe()

	// 3. Pair child device via QR scan (empty avatar -> gets default)
	dev, err := groupUC.PairDeviceFromQR(ctx, parentID, group.ID, QRScanPayload{
		DeviceUUID: "child-phone-uuid-1",
		DeviceName: "HP Kakak (Xiaomi)",
		IPv4:       "192.168.1.10",
		IPv6:       "fe80::2",
	})
	if err != nil {
		t.Fatalf("pair device failed: %v", err)
	}
	if dev.DeviceName != "HP Kakak (Xiaomi)" {
		t.Errorf("expected device name HP Kakak (Xiaomi), got %s", dev.DeviceName)
	}
	if !strings.HasPrefix(dev.Avatar, "char_") {
		t.Errorf("expected default avatar starting with char_, got %s", dev.Avatar)
	}

	// 4. Verify SSE event MemberJoined was received
	select {
	case evt := <-sseCh:
		if evt.Event != "MemberJoined" {
			t.Errorf("expected event MemberJoined, got %s", evt.Event)
		}
	case <-time.After(500 * time.Millisecond):
		t.Fatal("timeout waiting for MemberJoined SSE event")
	}

	// 5. Get group details
	gDetails, devices, err := groupUC.GetGroupDetails(ctx, parentID, group.ID)
	if err != nil {
		t.Fatalf("get group details failed: %v", err)
	}
	if gDetails.Name != "Keluarga Cemara" || len(devices) != 1 {
		t.Errorf("expected group Keluarga Cemara with 1 device, got %s with %d devices", gDetails.Name, len(devices))
	}
}

func TestTrackingAndSyncUseCase(t *testing.T) {
	ctx := context.Background()
	authUC, groupUC, trackingUC, syncUC, _, hub := setupTestEnv(t)

	authRes, err := authUC.Register(ctx, RegisterInput{
		Username: "ibu_bahagia", Email: "ibu@fam.id", PhoneNumber: "082222222222", Password: "password123",
	})
	if err != nil || authRes == nil {
		t.Fatalf("register failed in test: %v", err)
	}
	group, _ := groupUC.CreateGroup(ctx, authRes.User.ID, "Keluarga Bahagia")
	dev, _ := groupUC.PairDeviceFromQR(ctx, authRes.User.ID, group.ID, QRScanPayload{
		DeviceUUID: "child-uuid-2",
		DeviceName: "HP Adik",
	})

	sseCh, unsubscribe := hub.Subscribe(group.ID)
	defer unsubscribe()

	// 1. Live ping
	baseTime := time.Now().UnixMilli()
	ping, err := trackingUC.IngestLivePing(ctx, LocationPingInput{
		DeviceID:     dev.ID,
		GroupID:      group.ID,
		Latitude:     -6.2088,
		Longitude:    106.8456,
		Accuracy:     3.5,
		Speed:        10.2,
		Heading:      90.0,
		BatteryLevel: 92,
		RecordedAt:   baseTime - 60000, // 60 detik lalu
	})
	if err != nil {
		t.Fatalf("ingest live ping failed: %v", err)
	}
	if ping.Latitude != -6.2088 {
		t.Errorf("expected lat -6.2088, got %f", ping.Latitude)
	}

	// Verify SSE LocationUpdated event
	select {
	case evt := <-sseCh:
		if evt.Event != "LocationUpdated" {
			t.Errorf("expected event LocationUpdated, got %s", evt.Event)
		}
	case <-time.After(500 * time.Millisecond):
		t.Fatal("timeout waiting for LocationUpdated SSE event")
	}

	// 2. Offline batch sync
	syncRes, err := syncUC.SyncBatchOffline(ctx, group.ID, []LocationPingInput{
		{DeviceID: dev.ID, GroupID: group.ID, Latitude: -6.2100, Longitude: 106.8500, RecordedAt: baseTime - 30000},
		{DeviceID: dev.ID, GroupID: group.ID, Latitude: -6.2110, Longitude: 106.8510, RecordedAt: baseTime},
	})
	if err != nil {
		t.Fatalf("sync batch offline failed: %v", err)
	}
	if syncRes.TotalInserted != 2 {
		t.Errorf("expected 2 points inserted, got %d", syncRes.TotalInserted)
	}

	// 3. Query latest locations
	latestLocs, err := trackingUC.GetGroupLatestLocations(ctx, group.ID)
	if err != nil {
		t.Fatalf("get latest locations failed: %v", err)
	}
	if len(latestLocs) != 1 {
		t.Fatalf("expected 1 latest location, got %d", len(latestLocs))
	}
	if latestLocs[0].Latitude != -6.2110 {
		t.Errorf("expected latest latitude -6.2110, got %f", latestLocs[0].Latitude)
	}

	// 4. Query history
	history, err := trackingUC.GetDeviceHistory(ctx, dev.ID, 10)
	if err != nil {
		t.Fatalf("get history failed: %v", err)
	}
	if len(history) != 3 {
		t.Errorf("expected 3 history points, got %d", len(history))
	}
}

func TestAvatarUseCase(t *testing.T) {
	ctx := context.Background()
	authUC, groupUC, _, _, avatarUC, hub := setupTestEnv(t)

	authRes, err := authUC.Register(ctx, RegisterInput{
		Username: "bunda_cemara", Email: "bunda@fam.id", PhoneNumber: "083333333333", Password: "password123",
	})
	if err != nil || authRes == nil {
		t.Fatalf("register failed in test: %v", err)
	}
	group, _ := groupUC.CreateGroup(ctx, authRes.User.ID, "Fam Cemara")
	dev, _ := groupUC.PairDeviceFromQR(ctx, authRes.User.ID, group.ID, QRScanPayload{
		DeviceUUID: "child-dev-3",
		DeviceName: "Anak Bungsu",
	})

	sseCh, unsubscribe := hub.Subscribe(group.ID)
	defer unsubscribe()

	tempDir := t.TempDir()

	// 1. Upload User Avatar (Parent)
	dummyImageContent := "fake-png-binary-content-12345"
	userAvatarURL, err := avatarUC.UploadUserAvatar(
		ctx,
		authRes.User.ID,
		strings.NewReader(dummyImageContent),
		"my_selfie.png",
		tempDir,
	)
	if err != nil {
		t.Fatalf("upload user avatar failed: %v", err)
	}
	if !strings.HasPrefix(userAvatarURL, "/uploads/avatars/user_") {
		t.Errorf("expected user avatar URL starting with /uploads/avatars/user_, got %s", userAvatarURL)
	}

	// 2. Upload Device Avatar (Child)
	devAvatarURL, err := avatarUC.UploadDeviceAvatar(
		ctx,
		dev.ID,
		group.ID,
		strings.NewReader(dummyImageContent),
		"child_photo.jpg",
		tempDir,
	)
	if err != nil {
		t.Fatalf("upload device avatar failed: %v", err)
	}
	if !strings.HasPrefix(devAvatarURL, "/uploads/avatars/dev_") {
		t.Errorf("expected device avatar URL starting with /uploads/avatars/dev_, got %s", devAvatarURL)
	}

	// 3. Verify SSE MemberAvatarUpdated event was broadcasted
	select {
	case evt := <-sseCh:
		if evt.Event != "MemberAvatarUpdated" {
			t.Errorf("expected event MemberAvatarUpdated, got %s", evt.Event)
		}
	case <-time.After(500 * time.Millisecond):
		t.Fatal("timeout waiting for MemberAvatarUpdated SSE event")
	}

	// 4. Invalid file extension test
	_, err = avatarUC.UploadUserAvatar(
		ctx,
		authRes.User.ID,
		strings.NewReader("script-content"),
		"malicious.sh",
		tempDir,
	)
	if err != ErrInvalidImageType {
		t.Fatalf("expected ErrInvalidImageType, got %v", err)
	}
}
