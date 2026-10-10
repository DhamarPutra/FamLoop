package usecase

import (
	"context"
	"fmt"

	"famloop/internal/domain"
)

type SyncBatchResult struct {
	TotalReceived int `json:"total_received"`
	TotalInserted int `json:"total_inserted"`
}

type SyncUseCase struct {
	locRepo     domain.LocationRepository
	groupRepo   domain.GroupRepository
	broadcaster domain.EventBroadcaster
}

func NewSyncUseCase(
	locRepo domain.LocationRepository,
	groupRepo domain.GroupRepository,
	broadcaster domain.EventBroadcaster,
) *SyncUseCase {
	return &SyncUseCase{
		locRepo:     locRepo,
		groupRepo:   groupRepo,
		broadcaster: broadcaster,
	}
}

func (uc *SyncUseCase) SyncBatchOffline(ctx context.Context, groupID string, inputs []LocationPingInput) (*SyncBatchResult, error) {
	if len(inputs) == 0 {
		return &SyncBatchResult{TotalReceived: 0, TotalInserted: 0}, nil
	}

	validPoints := make([]domain.LocationPoint, 0, len(inputs))
	latestPointPerDevice := make(map[string]LocationPingInput)

	for _, in := range inputs {
		if in.DeviceID == "" {
			continue
		}
		if in.GroupID == "" {
			in.GroupID = groupID
		}
		// Validasi batas koordinat dasar
		if in.Latitude < -90 || in.Latitude > 90 || in.Latitude == 0 {
			continue
		}
		if in.Longitude < -180 || in.Longitude > 180 || in.Longitude == 0 {
			continue
		}

		validPoints = append(validPoints, domain.LocationPoint{
			DeviceID:   in.DeviceID,
			GroupID:    in.GroupID,
			Latitude:   in.Latitude,
			Longitude:  in.Longitude,
			Accuracy:   in.Accuracy,
			Speed:      in.Speed,
			Heading:    in.Heading,
			RecordedAt: in.RecordedAt,
		})

		// Lacak titik terbaru per perangkat untuk update heartbeat
		currentLatest, exists := latestPointPerDevice[in.DeviceID]
		if !exists || in.RecordedAt > currentLatest.RecordedAt {
			latestPointPerDevice[in.DeviceID] = in
		}
	}

	insertedCount, err := uc.locRepo.SaveBatch(ctx, validPoints)
	if err != nil {
		return nil, fmt.Errorf("failed to process offline sync batch: %w", err)
	}

	// Update heartbeat & siarkan titik paling mutakhir untuk tiap perangkat
	for deviceID, latestIn := range latestPointPerDevice {
		if latestIn.BatteryLevel > 0 {
			_ = uc.groupRepo.UpdateDeviceHeartbeat(ctx, deviceID, latestIn.BatteryLevel)
		} else {
			_ = uc.groupRepo.UpdateDeviceHeartbeat(ctx, deviceID, 100)
		}

		var deviceName, avatar string
		if dev, err := uc.groupRepo.GetDeviceByID(ctx, deviceID); err == nil && dev != nil {
			deviceName = dev.DeviceName
			avatar = dev.Avatar
		}

		if uc.broadcaster != nil {
			uc.broadcaster.BroadcastToGroup(latestIn.GroupID, "LocationUpdated", map[string]any{
				"device_id":     deviceID,
				"group_id":      latestIn.GroupID,
				"device_name":   deviceName,
				"avatar":        avatar,
				"latitude":      latestIn.Latitude,
				"longitude":     latestIn.Longitude,
				"accuracy":      latestIn.Accuracy,
				"speed":         latestIn.Speed,
				"heading":       latestIn.Heading,
				"battery_level": latestIn.BatteryLevel,
				"recorded_at":   latestIn.RecordedAt,
				"source":        "offline_sync",
			})
		}
	}

	return &SyncBatchResult{
		TotalReceived: len(inputs),
		TotalInserted: insertedCount,
	}, nil
}
