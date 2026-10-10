package main

import (
	"log"
	"os"
	"os/signal"
	"syscall"

	"famloop/internal/repository/sqlite"
	famHttp "famloop/internal/transport/http"
	"famloop/internal/transport/mqtt"
	"famloop/internal/transport/sse"
	"famloop/internal/usecase"

	"github.com/gofiber/fiber/v2"
)

func main() {
	port := getEnv("PORT", "8080")
	mqttPort := getEnv("MQTT_PORT", "1883")
	dbPath := getEnv("DB_PATH", "./database/famloop.db")
	jwtSecret := getEnv("JWT_SECRET", "famloop-secure-jwt-secret-key-32chars!")
	uploadDir := getEnv("UPLOAD_DIR", "./uploads")
	webDistDir := getEnv("WEB_DIST_DIR", "./web/dist")

	// 1. Inisialisasi direktori upload foto profil
	_ = os.MkdirAll(uploadDir+"/avatars", 0755)

	// 2. Inisialisasi Database SQLite WAL mode
	db, err := sqlite.NewDB(dbPath)
	if err != nil {
		log.Fatalf("[FATAL] Gagal menghubungkan ke database SQLite: %v", err)
	}
	defer db.Close()
	log.Printf("[INFO] Database SQLite siap di %s\n", dbPath)

	// 3. Inisialisasi Repositori
	userRepo := sqlite.NewUserRepository(db)
	groupRepo := sqlite.NewGroupRepository(db)
	locRepo := sqlite.NewLocationRepository(db)

	// 4. Inisialisasi SSE Hub (Broadcaster)
	sseHub := sse.NewHub()

	// 5. Inisialisasi Use Cases
	authUC := usecase.NewAuthUseCase(userRepo, jwtSecret)
	groupUC := usecase.NewGroupUseCase(groupRepo, sseHub)
	trackingUC := usecase.NewTrackingUseCase(locRepo, groupRepo, sseHub)
	syncUC := usecase.NewSyncUseCase(locRepo, groupRepo, sseHub)
	avatarUC := usecase.NewAvatarUseCase(userRepo, groupRepo, sseHub)

	// 6. Jalankan Embedded MQTT Broker
	mqttBroker := mqtt.NewBroker(trackingUC, mqttPort)
	if err := mqttBroker.Start(); err != nil {
		log.Printf("[WARN] Gagal menyalakan MQTT broker: %v (lanjut tanpa MQTT)\n", err)
	} else {
		defer mqttBroker.Close()
	}

	// 7. Inisialisasi Fiber HTTP App
	app := fiber.New(fiber.Config{
		ServerHeader: "FamLoop-Server",
		AppName:      "FamLoop Family Tracking v1.0",
		ErrorHandler: famHttp.NewErrorHandler(),
	})

	famHttp.SetupCommonMiddlewares(app)

	// 8. Inisialisasi HTTP Handlers & Routes
	authHandler := famHttp.NewAuthHandler(authUC, avatarUC, uploadDir)
	groupHandler := famHttp.NewGroupHandler(groupUC, avatarUC, uploadDir)
	trackingHandler := famHttp.NewTrackingHandler(trackingUC, syncUC, sseHub)

	famHttp.SetupRoutes(app, famHttp.RouterConfig{
		AuthHandler:     authHandler,
		GroupHandler:    groupHandler,
		TrackingHandler: trackingHandler,
		AuthUC:          authUC,
		UploadDir:       uploadDir,
		WebDistDir:      webDistDir,
	})

	// 9. Graceful Shutdown listener
	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, os.Interrupt, syscall.SIGTERM)

	go func() {
		<-sigCh
		log.Println("[INFO] Menerima sinyal shutdown, mematikan server secara graceful...")
		_ = app.Shutdown()
	}()

	log.Printf("[INFO] FamLoop HTTP Server berjalan di port :%s\n", port)
	if err := app.Listen(":" + port); err != nil {
		log.Printf("[INFO] Server berhenti: %v\n", err)
	}
}

func getEnv(key, defaultVal string) string {
	if val := os.Getenv(key); val != "" {
		return val
	}
	return defaultVal
}
