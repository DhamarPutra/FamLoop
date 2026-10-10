package http

import (
	"bufio"
	"encoding/json"
	"fmt"
	"strconv"
	"time"

	"famloop/internal/transport/sse"
	"famloop/internal/usecase"

	"github.com/gofiber/fiber/v2"
)

type TrackingHandler struct {
	trackingUC *usecase.TrackingUseCase
	syncUC     *usecase.SyncUseCase
	hub        *sse.Hub
}

func NewTrackingHandler(
	trackingUC *usecase.TrackingUseCase,
	syncUC *usecase.SyncUseCase,
	hub *sse.Hub,
) *TrackingHandler {
	return &TrackingHandler{
		trackingUC: trackingUC,
		syncUC:     syncUC,
		hub:        hub,
	}
}

func (h *TrackingHandler) Ping(c *fiber.Ctx) error {
	var input usecase.LocationPingInput
	if err := c.BodyParser(&input); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": fiber.Map{"code": fiber.StatusBadRequest, "message": "Payload JSON lokasi tidak valid"},
		})
	}

	point, err := h.trackingUC.IngestLivePing(c.Context(), input)
	if err != nil {
		return c.Status(fiber.StatusUnprocessableEntity).JSON(fiber.Map{
			"error": fiber.Map{"code": fiber.StatusUnprocessableEntity, "message": err.Error()},
		})
	}

	return c.JSON(fiber.Map{
		"ok":   true,
		"data": point,
	})
}

func (h *TrackingHandler) SyncBatch(c *fiber.Ctx) error {
	groupID := c.Query("group_id")

	// Payload bisa berupa array langsung [...] atau object {"group_id": "...", "points": [...]}
	var payload struct {
		GroupID string                    `json:"group_id"`
		Points  []usecase.LocationPingInput `json:"points"`
	}

	if err := c.BodyParser(&payload); err == nil && len(payload.Points) > 0 {
		if payload.GroupID != "" {
			groupID = payload.GroupID
		}
		result, err := h.syncUC.SyncBatchOffline(c.Context(), groupID, payload.Points)
		if err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"error": fiber.Map{"code": fiber.StatusInternalServerError, "message": err.Error()},
			})
		}
		return c.JSON(fiber.Map{"ok": true, "data": result})
	}

	// Coba parsing sebagai raw array
	var rawPoints []usecase.LocationPingInput
	if err := c.BodyParser(&rawPoints); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": fiber.Map{"code": fiber.StatusBadRequest, "message": "Format batch payload tidak valid"},
		})
	}

	result, err := h.syncUC.SyncBatchOffline(c.Context(), groupID, rawPoints)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": fiber.Map{"code": fiber.StatusInternalServerError, "message": err.Error()},
		})
	}

	return c.JSON(fiber.Map{"ok": true, "data": result})
}

func (h *TrackingHandler) GetLatestLocations(c *fiber.Ctx) error {
	groupID := c.Params("id")
	if groupID == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": fiber.Map{"code": fiber.StatusBadRequest, "message": "group id parameter diperlukan"},
		})
	}

	locations, err := h.trackingUC.GetGroupLatestLocations(c.Context(), groupID)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": fiber.Map{"code": fiber.StatusInternalServerError, "message": err.Error()},
		})
	}

	return c.JSON(fiber.Map{
		"data": locations,
	})
}

func (h *TrackingHandler) GetDeviceHistory(c *fiber.Ctx) error {
	deviceID := c.Params("device_id")
	limitStr := c.Query("limit", "100")
	limit, _ := strconv.Atoi(limitStr)

	history, err := h.trackingUC.GetDeviceHistory(c.Context(), deviceID, limit)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": fiber.Map{"code": fiber.StatusInternalServerError, "message": err.Error()},
		})
	}

	return c.JSON(fiber.Map{
		"data": history,
	})
}

func (h *TrackingHandler) GroupSSE(c *fiber.Ctx) error {
	groupID := c.Params("id")
	if groupID == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": fiber.Map{"code": fiber.StatusBadRequest, "message": "group id parameter diperlukan"},
		})
	}

	c.Set("Content-Type", "text/event-stream")
	c.Set("Cache-Control", "no-cache")
	c.Set("Connection", "keep-alive")
	c.Set("Transfer-Encoding", "chunked")
	c.Set("X-Accel-Buffering", "no")

	c.Context().SetBodyStreamWriter(func(w *bufio.Writer) {
		_ = writeSSEComment(w, "connected to famloop group sse")

		// 1. Kirim posisi terakhir seluruh anggota seketika saat client terhubung
		if latest, err := h.trackingUC.GetGroupLatestLocations(c.Context(), groupID); err == nil && len(latest) > 0 {
			latestBytes, _ := json.Marshal(latest)
			_ = writeSSEEvent(w, "LatestLocations", string(latestBytes))
		}

		// 2. Subscribe ke SSE Hub grup
		eventsCh, unsubscribe := h.hub.Subscribe(groupID)
		defer unsubscribe()

		ticker := time.NewTicker(15 * time.Second)
		defer ticker.Stop()

		for {
			select {
			case evt, ok := <-eventsCh:
				if !ok {
					return
				}
				dataBytes, err := json.Marshal(evt.Data)
				if err != nil {
					continue
				}
				if err := writeSSEEvent(w, evt.Event, string(dataBytes)); err != nil {
					return
				}

			case <-ticker.C:
				if err := writeSSEComment(w, "ping"); err != nil {
					return
				}
			}
		}
	})

	return nil
}

func writeSSEEvent(w *bufio.Writer, event string, data string) error {
	if event != "" {
		if _, err := fmt.Fprintf(w, "event: %s\n", event); err != nil {
			return err
		}
	}
	if _, err := fmt.Fprintf(w, "data: %s\n\n", data); err != nil {
		return err
	}
	return w.Flush()
}

func writeSSEComment(w *bufio.Writer, comment string) error {
	if _, err := fmt.Fprintf(w, ": %s\n\n", comment); err != nil {
		return err
	}
	return w.Flush()
}
