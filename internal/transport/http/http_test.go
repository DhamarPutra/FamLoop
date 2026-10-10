package http

import (
	"bytes"
	"encoding/json"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
	"time"

	"famloop/internal/repository/sqlite"
	"famloop/internal/transport/sse"
	"famloop/internal/usecase"

	"github.com/gofiber/fiber/v2"
)

func setupTestApp(t *testing.T) (*fiber.App, string) {
	db, err := sqlite.NewDB(":memory:")
	if err != nil {
		t.Fatalf("failed to init db: %v", err)
	}

	userRepo := sqlite.NewUserRepository(db)
	groupRepo := sqlite.NewGroupRepository(db)
	locRepo := sqlite.NewLocationRepository(db)
	sseHub := sse.NewHub()

	jwtSecret := "test-secret-key-for-http-tests!"
	authUC := usecase.NewAuthUseCase(userRepo, jwtSecret)
	groupUC := usecase.NewGroupUseCase(groupRepo, sseHub)
	trackingUC := usecase.NewTrackingUseCase(locRepo, groupRepo, sseHub)
	syncUC := usecase.NewSyncUseCase(locRepo, groupRepo, sseHub)
	avatarUC := usecase.NewAvatarUseCase(userRepo, groupRepo, sseHub)

	tempDir := t.TempDir()

	app := fiber.New(fiber.Config{
		ErrorHandler: NewErrorHandler(),
	})
	SetupCommonMiddlewares(app)

	authHandler := NewAuthHandler(authUC, avatarUC, tempDir)
	groupHandler := NewGroupHandler(groupUC, avatarUC, tempDir)
	trackingHandler := NewTrackingHandler(trackingUC, syncUC, sseHub)

	SetupRoutes(app, RouterConfig{
		AuthHandler:     authHandler,
		GroupHandler:    groupHandler,
		TrackingHandler: trackingHandler,
		AuthUC:          authUC,
		UploadDir:       tempDir,
		WebDistDir:      "../../../web/dist",
	})

	return app, jwtSecret
}

func TestHealthEndpoint(t *testing.T) {
	app, _ := setupTestApp(t)

	req := httptest.NewRequest("GET", "/health", nil)
	resp, err := app.Test(req)
	if err != nil {
		t.Fatalf("health request failed: %v", err)
	}

	if resp.StatusCode != http.StatusOK {
		t.Errorf("expected status 200, got %d", resp.StatusCode)
	}
}

func TestAstroLandingPageAndSEORoutes(t *testing.T) {
	st, err := os.Stat("../../../web/dist")
	if err != nil || !st.IsDir() {
		t.Skip("skipping Astro static route tests if web/dist not found")
	}
	app, _ := setupTestApp(t)

	// 1. Root landing page
	req := httptest.NewRequest("GET", "/", nil)
	resp, err := app.Test(req)
	if err != nil {
		t.Fatalf("request / failed: %v", err)
	}
	if resp.StatusCode != http.StatusOK {
		t.Errorf("expected status 200 for root landing page, got %d", resp.StatusCode)
	}

	// 2. Sitemap XML
	req = httptest.NewRequest("GET", "/sitemap.xml", nil)
	resp, _ = app.Test(req)
	if resp.StatusCode != http.StatusOK {
		t.Errorf("expected status 200 for sitemap.xml, got %d", resp.StatusCode)
	}

	// 3. AI Context file (llms.txt)
	req = httptest.NewRequest("GET", "/llms.txt", nil)
	resp, _ = app.Test(req)
	if resp.StatusCode != http.StatusOK {
		t.Errorf("expected status 200 for llms.txt, got %d", resp.StatusCode)
	}
}

func TestAuthAndGroupEndpoints(t *testing.T) {
	app, _ := setupTestApp(t)

	// 1. Register Parent
	regBody, _ := json.Marshal(map[string]string{
		"username":     "bapak_eko",
		"email":        "eko@famloop.id",
		"phone_number": "081234567890",
		"password":     "password123",
	})
	req := httptest.NewRequest("POST", "/api/v1/auth/register", bytes.NewReader(regBody))
	req.Header.Set("Content-Type", "application/json")
	resp, err := app.Test(req)
	if err != nil || resp.StatusCode != http.StatusCreated {
		t.Fatalf("register failed, status: %d", resp.StatusCode)
	}

	// 2. Login
	loginBody, _ := json.Marshal(map[string]string{
		"username_or_email": "eko@famloop.id",
		"password":          "password123",
	})
	req = httptest.NewRequest("POST", "/api/v1/auth/login", bytes.NewReader(loginBody))
	req.Header.Set("Content-Type", "application/json")
	resp, err = app.Test(req)
	if err != nil || resp.StatusCode != http.StatusOK {
		t.Fatalf("login failed, status: %d", resp.StatusCode)
	}

	var loginResp struct {
		Data struct {
			Token string `json:"token"`
		} `json:"data"`
	}
	_ = json.NewDecoder(resp.Body).Decode(&loginResp)
	token := loginResp.Data.Token
	if token == "" {
		t.Fatal("expected non-empty token")
	}

	// 3. Get Profile (Me)
	req = httptest.NewRequest("GET", "/api/v1/auth/me", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	resp, _ = app.Test(req)
	if resp.StatusCode != http.StatusOK {
		t.Errorf("expected 200 OK for /me, got %d", resp.StatusCode)
	}

	// 4. Create Family Group
	groupBody, _ := json.Marshal(map[string]string{
		"name": "Keluarga Pak Eko",
	})
	req = httptest.NewRequest("POST", "/api/v1/groups", bytes.NewReader(groupBody))
	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("Content-Type", "application/json")
	resp, _ = app.Test(req)
	if resp.StatusCode != http.StatusCreated {
		t.Fatalf("create group failed, status: %d", resp.StatusCode)
	}

	var groupResp struct {
		Data struct {
			ID   string `json:"id"`
			Name string `json:"name"`
		} `json:"data"`
	}
	_ = json.NewDecoder(resp.Body).Decode(&groupResp)
	groupID := groupResp.Data.ID

	// 5. Pair child device via QR code scan
	pairBody, _ := json.Marshal(map[string]string{
		"device_uuid": "android-hw-998877",
		"device_name": "HP Anak Pertama",
		"ipv4":        "192.168.1.15",
		"ipv6":        "fe80::100",
		"mac_address": "00:aa:bb:cc:dd:ee",
	})
	req = httptest.NewRequest("POST", "/api/v1/groups/"+groupID+"/pair-device", bytes.NewReader(pairBody))
	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("Content-Type", "application/json")
	resp, _ = app.Test(req)
	if resp.StatusCode != http.StatusCreated {
		t.Fatalf("pair device failed, status: %d", resp.StatusCode)
	}

	var pairResp struct {
		Data struct {
			ID     string `json:"id"`
			Avatar string `json:"avatar"`
		} `json:"data"`
	}
	_ = json.NewDecoder(resp.Body).Decode(&pairResp)
	deviceID := pairResp.Data.ID
	if deviceID == "" {
		t.Fatal("expected device ID from pairing")
	}

	// 6. Send Location Ping
	nowMs := time.Now().UnixMilli()
	pingBody, _ := json.Marshal(map[string]any{
		"device_id":     deviceID,
		"group_id":      groupID,
		"latitude":      -6.2000,
		"longitude":     106.8166,
		"accuracy":      5.0,
		"speed":         15.5,
		"battery_level": 88,
		"recorded_at":   nowMs - 20000,
	})
	req = httptest.NewRequest("POST", "/api/v1/tracking/ping", bytes.NewReader(pingBody))
	req.Header.Set("Content-Type", "application/json")
	resp, _ = app.Test(req)
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("ping failed, status: %d", resp.StatusCode)
	}

	// 7. Send Offline Batch Sync
	syncBody, _ := json.Marshal(map[string]any{
		"group_id": groupID,
		"points": []map[string]any{
			{"device_id": deviceID, "group_id": groupID, "latitude": -6.2010, "longitude": 106.8170, "recorded_at": nowMs - 10000},
			{"device_id": deviceID, "group_id": groupID, "latitude": -6.2020, "longitude": 106.8180, "recorded_at": nowMs},
		},
	})
	req = httptest.NewRequest("POST", "/api/v1/tracking/sync/batch", bytes.NewReader(syncBody))
	req.Header.Set("Content-Type", "application/json")
	resp, _ = app.Test(req)
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("batch sync failed, status: %d", resp.StatusCode)
	}

	// 8. Get Latest Locations for Map View
	req = httptest.NewRequest("GET", "/api/v1/groups/"+groupID+"/locations/latest", nil)
	resp, _ = app.Test(req)
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("get latest locations failed, status: %d", resp.StatusCode)
	}

	var latestResp struct {
		Data []struct {
			DeviceID  string  `json:"device_id"`
			Latitude  float64 `json:"latitude"`
			Longitude float64 `json:"longitude"`
			Avatar    string  `json:"avatar"`
		} `json:"data"`
	}
	_ = json.NewDecoder(resp.Body).Decode(&latestResp)
	if len(latestResp.Data) != 1 {
		t.Fatalf("expected 1 latest location, got %d", len(latestResp.Data))
	}
	if latestResp.Data[0].Latitude != -6.2020 {
		t.Errorf("expected latest latitude -6.2020, got %f", latestResp.Data[0].Latitude)
	}

	// 9. Get History
	req = httptest.NewRequest("GET", "/api/v1/groups/"+groupID+"/members/"+deviceID+"/history", nil)
	resp, _ = app.Test(req)
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("get history failed, status: %d", resp.StatusCode)
	}
}

func TestProfileUpdateAndDeviceEndpoints(t *testing.T) {
	app, _ := setupTestApp(t)

	// Register & Login Parent
	regBody, _ := json.Marshal(map[string]string{
		"username":     "ibu_budi",
		"email":        "budi@famloop.id",
		"phone_number": "081298765432",
		"password":     "password123",
	})
	req := httptest.NewRequest("POST", "/api/v1/auth/register", bytes.NewReader(regBody))
	req.Header.Set("Content-Type", "application/json")
	resp, err := app.Test(req)
	if err != nil || resp.StatusCode != http.StatusCreated {
		t.Fatalf("register failed: %v", err)
	}

	var regResp struct {
		Data struct {
			Token string `json:"token"`
		} `json:"data"`
	}
	_ = json.NewDecoder(resp.Body).Decode(&regResp)
	token := regResp.Data.Token

	// 1. Update Profile (PUT /api/v1/auth/me)
	updateBody, _ := json.Marshal(map[string]string{
		"username":     "ibu_budi_updated",
		"email":        "budi_new@famloop.id",
		"phone_number": "081299998888",
	})
	req = httptest.NewRequest("PUT", "/api/v1/auth/me", bytes.NewReader(updateBody))
	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("Content-Type", "application/json")
	resp, _ = app.Test(req)
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("expected 200 for update profile, got %d", resp.StatusCode)
	}

	var profileResp struct {
		Data struct {
			Username string `json:"username"`
			Email    string `json:"email"`
		} `json:"data"`
	}
	_ = json.NewDecoder(resp.Body).Decode(&profileResp)
	if profileResp.Data.Username != "ibu_budi_updated" || profileResp.Data.Email != "budi_new@famloop.id" {
		t.Errorf("profile fields not updated properly: %+v", profileResp.Data)
	}

	// 2. Create Group
	grpBody, _ := json.Marshal(map[string]string{"name": "Grup Budi"})
	req = httptest.NewRequest("POST", "/api/v1/groups", bytes.NewReader(grpBody))
	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("Content-Type", "application/json")
	resp, _ = app.Test(req)
	var grpResp struct {
		Data struct {
			ID string `json:"id"`
		} `json:"data"`
	}
	_ = json.NewDecoder(resp.Body).Decode(&grpResp)
	groupID := grpResp.Data.ID

	// 3. List Groups (GET /api/v1/groups)
	req = httptest.NewRequest("GET", "/api/v1/groups", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	resp, _ = app.Test(req)
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("expected 200 for list groups, got %d", resp.StatusCode)
	}

	// 4. Pair a Device
	pairBody, _ := json.Marshal(map[string]string{
		"device_uuid": "budi-child-phone-001",
		"device_name": "HP Budi Kecil",
		"ipv4":        "192.168.1.100",
	})
	req = httptest.NewRequest("POST", "/api/v1/groups/"+groupID+"/pair-device", bytes.NewReader(pairBody))
	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("Content-Type", "application/json")
	resp, _ = app.Test(req)
	var pairResp struct {
		Data struct {
			ID string `json:"id"`
		} `json:"data"`
	}
	_ = json.NewDecoder(resp.Body).Decode(&pairResp)
	deviceID := pairResp.Data.ID

	// 5. Query Device Details (GET /api/v1/devices/:id)
	req = httptest.NewRequest("GET", "/api/v1/devices/"+deviceID, nil)
	resp, _ = app.Test(req)
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("expected 200 for get device, got %d", resp.StatusCode)
	}

	// 6. Get Group Details with Members (GET /api/v1/groups/:id)
	req = httptest.NewRequest("GET", "/api/v1/groups/"+groupID, nil)
	req.Header.Set("Authorization", "Bearer "+token)
	resp, _ = app.Test(req)
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("expected 200 for get group details, got %d", resp.StatusCode)
	}
	var detailsResp struct {
		Data struct {
			Members []struct {
				ID string `json:"id"`
			} `json:"members"`
		} `json:"data"`
	}
	_ = json.NewDecoder(resp.Body).Decode(&detailsResp)
	if len(detailsResp.Data.Members) != 1 {
		t.Fatalf("expected 1 member, got %d", len(detailsResp.Data.Members))
	}

	// 7. Remove Member Device (DELETE /api/v1/groups/:id/members/:device_id)
	req = httptest.NewRequest("DELETE", "/api/v1/groups/"+groupID+"/members/"+deviceID, nil)
	req.Header.Set("Authorization", "Bearer "+token)
	resp, _ = app.Test(req)
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("expected 200 for remove member, got %d", resp.StatusCode)
	}

	// Verify member is removed from group
	req = httptest.NewRequest("GET", "/api/v1/groups/"+groupID, nil)
	req.Header.Set("Authorization", "Bearer "+token)
	resp, _ = app.Test(req)
	_ = json.NewDecoder(resp.Body).Decode(&detailsResp)
	if len(detailsResp.Data.Members) != 0 {
		t.Fatalf("expected 0 members after deletion, got %d", len(detailsResp.Data.Members))
	}

	// 8. Delete Group (DELETE /api/v1/groups/:id)
	req = httptest.NewRequest("DELETE", "/api/v1/groups/"+groupID, nil)
	req.Header.Set("Authorization", "Bearer "+token)
	resp, _ = app.Test(req)
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("expected 200 for delete group, got %d", resp.StatusCode)
	}

	// Verify group is 404
	req = httptest.NewRequest("GET", "/api/v1/groups/"+groupID, nil)
	req.Header.Set("Authorization", "Bearer "+token)
	resp, _ = app.Test(req)
	if resp.StatusCode != http.StatusNotFound {
		t.Fatalf("expected 404 after group deletion, got %d", resp.StatusCode)
	}
}

func TestAvatarUploadEndpoints(t *testing.T) {
	app, _ := setupTestApp(t)

	// Register & Login Parent
	regBody, _ := json.Marshal(map[string]string{
		"username":     "ayah_dani",
		"email":        "dani@famloop.id",
		"phone_number": "081211112222",
		"password":     "password123",
	})
	req := httptest.NewRequest("POST", "/api/v1/auth/register", bytes.NewReader(regBody))
	req.Header.Set("Content-Type", "application/json")
	resp, _ := app.Test(req)
	var regResp struct {
		Data struct {
			Token string `json:"token"`
		} `json:"data"`
	}
	_ = json.NewDecoder(resp.Body).Decode(&regResp)
	token := regResp.Data.Token

	// Helper function to build multipart file request
	createMultipartRequest := func(method, url string, fieldName, fileName string, fileContent []byte) *http.Request {
		body := &bytes.Buffer{}
		writer := multipart.NewWriter(body)
		part, _ := writer.CreateFormFile(fieldName, fileName)
		_, _ = part.Write(fileContent)
		_ = writer.Close()

		r := httptest.NewRequest(method, url, body)
		r.Header.Set("Content-Type", writer.FormDataContentType())
		return r
	}

	// 1. Upload Parent Avatar (POST /api/v1/auth/avatar)
	dummyImage := []byte("\x89PNG\r\n\x1a\n\x00\x00\x00\rIHDR\x00\x00\x00\x01\x00\x00\x00\x01\x08\x06\x00\x00\x00\x1f\x15c4")
	req = createMultipartRequest("POST", "/api/v1/auth/avatar", "avatar", "profile.png", dummyImage)
	req.Header.Set("Authorization", "Bearer "+token)
	resp, _ = app.Test(req)
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("expected 200 for upload parent avatar, got %d", resp.StatusCode)
	}

	var avatarResp struct {
		Data struct {
			AvatarURL string `json:"avatar_url"`
		} `json:"data"`
	}
	_ = json.NewDecoder(resp.Body).Decode(&avatarResp)
	if avatarResp.Data.AvatarURL == "" {
		t.Fatal("expected non-empty avatar_url")
	}

	// 2. Create Group and Pair Device
	grpBody, _ := json.Marshal(map[string]string{"name": "Grup Dani"})
	req = httptest.NewRequest("POST", "/api/v1/groups", bytes.NewReader(grpBody))
	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("Content-Type", "application/json")
	resp, _ = app.Test(req)
	var grpResp struct {
		Data struct {
			ID string `json:"id"`
		} `json:"data"`
	}
	_ = json.NewDecoder(resp.Body).Decode(&grpResp)
	groupID := grpResp.Data.ID

	pairBody, _ := json.Marshal(map[string]string{
		"device_uuid": "dani-child-dev",
		"device_name": "HP Dani Jr",
	})
	req = httptest.NewRequest("POST", "/api/v1/groups/"+groupID+"/pair-device", bytes.NewReader(pairBody))
	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("Content-Type", "application/json")
	resp, _ = app.Test(req)
	var pairResp struct {
		Data struct {
			ID string `json:"id"`
		} `json:"data"`
	}
	_ = json.NewDecoder(resp.Body).Decode(&pairResp)
	deviceID := pairResp.Data.ID

	// 3. Upload Child Device Avatar (POST /api/v1/devices/:id/avatar)
	req = createMultipartRequest("POST", "/api/v1/devices/"+deviceID+"/avatar", "avatar", "child.jpg", dummyImage)
	resp, _ = app.Test(req)
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("expected 200 for upload child device avatar, got %d", resp.StatusCode)
	}
}

func TestEdgeCasesAndValidation(t *testing.T) {
	app, _ := setupTestApp(t)

	// 1. Invalid Registration (missing fields)
	badReg, _ := json.Marshal(map[string]string{
		"username": "short",
	})
	req := httptest.NewRequest("POST", "/api/v1/auth/register", bytes.NewReader(badReg))
	req.Header.Set("Content-Type", "application/json")
	resp, _ := app.Test(req)
	if resp.StatusCode != http.StatusBadRequest {
		t.Errorf("expected 400 for bad register, got %d", resp.StatusCode)
	}

	// 2. Bad Login
	badLogin, _ := json.Marshal(map[string]string{
		"username_or_email": "nonexistent@user.com",
		"password":          "wrongpassword",
	})
	req = httptest.NewRequest("POST", "/api/v1/auth/login", bytes.NewReader(badLogin))
	req.Header.Set("Content-Type", "application/json")
	resp, _ = app.Test(req)
	if resp.StatusCode != http.StatusUnauthorized {
		t.Errorf("expected 401 for wrong credentials, got %d", resp.StatusCode)
	}

	// 3. Unauthorized Access to Protected Route
	req = httptest.NewRequest("GET", "/api/v1/auth/me", nil)
	resp, _ = app.Test(req)
	if resp.StatusCode != http.StatusUnauthorized {
		t.Errorf("expected 401 without bearer token, got %d", resp.StatusCode)
	}

	// 4. Invalid Coordinates in Tracking Ping (lat > 90)
	invalidPing, _ := json.Marshal(map[string]any{
		"device_id":   "dev-123",
		"group_id":    "grp-123",
		"latitude":    195.0,
		"longitude":   106.8,
		"recorded_at": time.Now().UnixMilli(),
	})
	req = httptest.NewRequest("POST", "/api/v1/tracking/ping", bytes.NewReader(invalidPing))
	req.Header.Set("Content-Type", "application/json")
	resp, _ = app.Test(req)
	if resp.StatusCode != http.StatusUnprocessableEntity {
		t.Errorf("expected 422 for out-of-range latitude, got %d", resp.StatusCode)
	}
}
