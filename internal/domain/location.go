package domain

import "time"

type LocationPoint struct {
	ID         int64     `json:"id,omitempty"`
	DeviceID   string    `json:"device_id"`
	GroupID    string    `json:"group_id"`
	Latitude   float64   `json:"latitude"`
	Longitude  float64   `json:"longitude"`
	Accuracy   float64   `json:"accuracy"`
	Speed      float64   `json:"speed"`
	Heading    float64   `json:"heading"`
	RecordedAt int64     `json:"recorded_at"` // Epoch in milliseconds
	SyncedAt   time.Time `json:"synced_at,omitempty"`
}

type MemberLatestLocation struct {
	DeviceID     string    `json:"device_id"`
	DeviceUUID   string    `json:"device_uuid"`
	DeviceName   string    `json:"device_name"`
	Avatar       string    `json:"avatar"`
	BatteryLevel int       `json:"battery_level"`
	IsActive     bool      `json:"is_active"`
	Latitude     float64   `json:"latitude"`
	Longitude    float64   `json:"longitude"`
	Accuracy     float64   `json:"accuracy"`
	Speed        float64   `json:"speed"`
	Heading      float64   `json:"heading"`
	RecordedAt   int64     `json:"recorded_at"`
	LastSeenAt   time.Time `json:"last_seen_at"`
}
