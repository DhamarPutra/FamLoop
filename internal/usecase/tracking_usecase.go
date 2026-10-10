package usecase

import (
	"context"
	"errors"
	"fmt"
	"time"

	"famloop/internal/domain"
)

var (
	ErrInvalidCoordinates = errors.New("koordinat latitude (-90 s/d 90) atau longitude (-180 s/d 180) tidak valid")
	ErrMissingDeviceOrGroup = errors.New("device_id dan group_id wajib diisi")
)

type LocationPingInput struct {
	DeviceID     string  `json:"device_id"`
	GroupID      string  `json:"group_id"`
	Latitude     float64 `json:"latitude"`
	Longitude    float64 `json:"longitude"`
	Accuracy     float64 `json:"accuracy"`
	Speed        float64 `json:"speed"`
	Heading      float64 `json:"heading"`
	BatteryLevel int     `json:"battery_level,omitempty"`
	RecordedAt   int64   `json:"recorded_at,omitempty"`
}

type TrackingUseCase struct {
	locRepo     domain.LocationRepository
	groupRepo   domain.GroupRepository
	broadcaster domain.EventBroadcaster
}

func NewTrackingUseCase(
	locRepo domain.LocationRepository,
	groupRepo domain.GroupRepository,
	broadcaster domain.EventBroadcaster,
) *TrackingUseCase {
	return &TrackingUseCase{
		locRepo:     locRepo,
		groupRepo:   groupRepo,
		broadcaster: broadcaster,
	}
}

func (uc *TrackingUseCase) IngestLivePing(ctx context.Context, input LocationPingInput) (*domain.LocationPoint, error) {
	if input.DeviceID == "" || input.GroupID == "" {
		return nil, ErrMissingDeviceOrGroup
	}

	if input.Latitude < -90 || input.Latitude > 90 || input.Latitude == 0 {
		return nil, ErrInvalidCoordinates
	}
	if input.Longitude < -180 || input.Longitude > 180 || input.Longitude == 0 {
		return nil, ErrInvalidCoordinates
	}

	if input.RecordedAt <= 0 {
		input.RecordedAt = time.Now().UnixMilli()
	}

	point := &domain.LocationPoint{
		DeviceID:   input.DeviceID,
		GroupID:    input.GroupID,
		Latitude:   input.Latitude,
		Longitude:  input.Longitude,
		Accuracy:   input.Accuracy,
		Speed:      input.Speed,
		Heading:    input.Heading,
		RecordedAt: input.RecordedAt,
	}

	if err := uc.locRepo.SavePoint(ctx, point); err != nil {
		return nil, fmt.Errorf("failed to save location: %w", err)
	}

	// Update last_seen_at dan battery perangkat
	if input.BatteryLevel > 0 {
		_ = uc.groupRepo.UpdateDeviceHeartbeat(ctx, input.DeviceID, input.BatteryLevel)
	} else {
		_ = uc.groupRepo.UpdateDeviceHeartbeat(ctx, input.DeviceID, 100)
	}

	// Ambil detail device untuk broadcast avatar dan nama ke peta
	var deviceName, avatar string
	if dev, err := uc.groupRepo.GetDeviceByID(ctx, input.DeviceID); err == nil && dev != nil {
		deviceName = dev.DeviceName
		avatar = dev.Avatar
	}

	// Siarkan event update real-time ke seluruh koneksi SSE grup
	if uc.broadcaster != nil {
		uc.broadcaster.BroadcastToGroup(input.GroupID, "LocationUpdated", map[string]any{
			"device_id":     input.DeviceID,
			"group_id":      input.GroupID,
			"device_name":   deviceName,
			"avatar":        avatar,
			"latitude":      input.Latitude,
			"longitude":     input.Longitude,
			"accuracy":      input.Accuracy,
			"speed":         input.Speed,
			"heading":       input.Heading,
			"battery_level": input.BatteryLevel,
			"recorded_at":   input.RecordedAt,
		})
	}

	return point, nil
}

func (uc *TrackingUseCase) GetGroupLatestLocations(ctx context.Context, groupID string) ([]domain.MemberLatestLocation, error) {
	return uc.locRepo.GetLatestGroupLocations(ctx, groupID)
}

func (uc *TrackingUseCase) GetDeviceHistory(ctx context.Context, deviceID string, limit int) ([]domain.LocationPoint, error) {
	return uc.locRepo.GetDeviceHistory(ctx, deviceID, limit)
}
