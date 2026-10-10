package mqtt

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"log"
	"strings"
	"sync"

	mqtt "github.com/mochi-mqtt/server/v2"
	"github.com/mochi-mqtt/server/v2/hooks/auth"
	"github.com/mochi-mqtt/server/v2/listeners"
	"github.com/mochi-mqtt/server/v2/packets"

	"famloop/internal/usecase"
)

type Broker struct {
	server     *mqtt.Server
	trackingUC *usecase.TrackingUseCase
	port       string
	closeOnce  sync.Once
}

func NewBroker(trackingUC *usecase.TrackingUseCase, port string) *Broker {
	if port == "" {
		port = "1883"
	}
	return &Broker{
		trackingUC: trackingUC,
		port:       port,
	}
}

func (b *Broker) Start() error {
	b.server = mqtt.New(nil)

	// Izinkan koneksi client
	_ = b.server.AddHook(new(auth.AllowHook), nil)

	// Tambahkan hook listener untuk topic tracking
	_ = b.server.AddHook(&trackingHook{trackingUC: b.trackingUC}, nil)

	tcp := listeners.NewTCP(listeners.Config{
		ID:      "famloop-tcp",
		Address: ":" + b.port,
	})

	if err := b.server.AddListener(tcp); err != nil {
		return fmt.Errorf("failed to add mqtt tcp listener: %w", err)
	}

	go func() {
		log.Printf("[INFO] FamLoop Embedded MQTT Broker running on :%s\n", b.port)
		if err := b.server.Serve(); err != nil {
			log.Printf("[WARN] MQTT server error: %v\n", err)
		}
	}()

	return nil
}

func (b *Broker) Close() error {
	var err error
	b.closeOnce.Do(func() {
		if b.server != nil {
			err = b.server.Close()
		}
	})
	return err
}

type trackingHook struct {
	mqtt.HookBase
	trackingUC *usecase.TrackingUseCase
}

func (h *trackingHook) ID() string {
	return "famloop-tracking-hook"
}

func (h *trackingHook) Provides(b byte) bool {
	return bytes.Contains([]byte{
		mqtt.OnPublish,
	}, []byte{b})
}

func (h *trackingHook) OnPublish(cl *mqtt.Client, pk packets.Packet) (packets.Packet, error) {
	topic := pk.TopicName
	// Format topic: famloop/devices/{device_id}/location atau famloop/tracking
	if strings.HasPrefix(topic, "famloop/devices/") || topic == "famloop/tracking" {
		var ping usecase.LocationPingInput
		if err := json.Unmarshal(pk.Payload, &ping); err == nil && ping.DeviceID != "" && ping.GroupID != "" {
			go func() {
				_, _ = h.trackingUC.IngestLivePing(context.Background(), ping)
			}()
		}
	}
	return pk, nil
}
