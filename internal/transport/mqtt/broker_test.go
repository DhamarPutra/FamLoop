package mqtt

import (
	"testing"
	"time"

	"famloop/internal/repository/sqlite"
	"famloop/internal/transport/sse"
	"famloop/internal/usecase"
)

func TestMQTTBrokerLifecycle(t *testing.T) {
	db, err := sqlite.NewDB(":memory:")
	if err != nil {
		t.Fatalf("failed to init db: %v", err)
	}
	locRepo := sqlite.NewLocationRepository(db)
	groupRepo := sqlite.NewGroupRepository(db)
	sseHub := sse.NewHub()
	trackingUC := usecase.NewTrackingUseCase(locRepo, groupRepo, sseHub)

	// Port 18883 to avoid conflict with default 1883
	broker := NewBroker(trackingUC, "18883")
	err = broker.Start()
	if err != nil {
		t.Skipf("skipping MQTT broker test if port cannot be bound: %v", err)
		return
	}
	defer broker.Close()

	// Wait briefly for serve goroutine
	time.Sleep(100 * time.Millisecond)

	if err := broker.Close(); err != nil {
		t.Errorf("failed to close broker: %v", err)
	}
}
