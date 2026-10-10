package domain

import "time"

type MemberDevice struct {
	ID           string     `json:"id"`
	GroupID      string     `json:"group_id"`
	DeviceUUID   string     `json:"device_uuid"`
	DeviceName   string     `json:"device_name"`
	Avatar       string     `json:"avatar"`
	IPv4         string     `json:"ipv4,omitempty"`
	IPv6         string     `json:"ipv6,omitempty"`
	MACAddress   string     `json:"mac_address,omitempty"`
	BatteryLevel int        `json:"battery_level"`
	IsActive     bool       `json:"is_active"`
	PairedAt     time.Time  `json:"paired_at"`
	LastSeenAt   *time.Time `json:"last_seen_at,omitempty"`
}
