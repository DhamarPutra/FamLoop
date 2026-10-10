package sqlite

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	"famloop/internal/domain"
)

type GroupRepository struct {
	db *sql.DB
}

func NewGroupRepository(db *sql.DB) *GroupRepository {
	return &GroupRepository{db: db}
}

func (r *GroupRepository) Create(ctx context.Context, g *domain.FamilyGroup) error {
	query := `INSERT INTO family_groups (id, name, parent_id, invite_code, created_at) VALUES (?, ?, ?, ?, ?)`
	g.CreatedAt = time.Now().UTC()
	_, err := r.db.ExecContext(ctx, query, g.ID, g.Name, g.ParentID, g.InviteCode, g.CreatedAt)
	if err != nil {
		return fmt.Errorf("failed to insert group: %w", err)
	}
	return nil
}

func (r *GroupRepository) GetByID(ctx context.Context, id string) (*domain.FamilyGroup, error) {
	query := `SELECT id, name, parent_id, invite_code, created_at FROM family_groups WHERE id = ?`
	row := r.db.QueryRowContext(ctx, query, id)

	var g domain.FamilyGroup
	var inviteCode sql.NullString
	err := row.Scan(&g.ID, &g.Name, &g.ParentID, &inviteCode, &g.CreatedAt)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, fmt.Errorf("scan group failed: %w", err)
	}
	if inviteCode.Valid {
		g.InviteCode = inviteCode.String
	}
	return &g, nil
}

func (r *GroupRepository) GetByParentID(ctx context.Context, parentID string) ([]domain.FamilyGroup, error) {
	query := `SELECT id, name, parent_id, invite_code, created_at FROM family_groups WHERE parent_id = ? ORDER BY created_at DESC`
	rows, err := r.db.QueryContext(ctx, query, parentID)
	if err != nil {
		return nil, fmt.Errorf("query groups failed: %w", err)
	}
	defer rows.Close()

	var groups []domain.FamilyGroup
	for rows.Next() {
		var g domain.FamilyGroup
		var inviteCode sql.NullString
		if err := rows.Scan(&g.ID, &g.Name, &g.ParentID, &inviteCode, &g.CreatedAt); err != nil {
			return nil, fmt.Errorf("scan group item failed: %w", err)
		}
		if inviteCode.Valid {
			g.InviteCode = inviteCode.String
		}
		groups = append(groups, g)
	}
	return groups, nil
}

func (r *GroupRepository) DeleteGroup(ctx context.Context, groupID string) error {
	query := `DELETE FROM family_groups WHERE id = ?`
	_, err := r.db.ExecContext(ctx, query, groupID)
	if err != nil {
		return fmt.Errorf("failed to delete group: %w", err)
	}
	return nil
}

func (r *GroupRepository) AddMemberDevice(ctx context.Context, d *domain.MemberDevice) error {
	if d.Avatar == "" {
		d.Avatar = domain.GetRandomDefaultAvatar()
	}

	query := `
		INSERT INTO member_devices (id, group_id, device_uuid, device_name, avatar, ipv4, ipv6, mac_address, battery_level, is_active, paired_at, last_seen_at)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, 1, ?, ?)
		ON CONFLICT(group_id, device_uuid) DO UPDATE SET
			device_name = excluded.device_name,
			avatar = CASE WHEN excluded.avatar != '' THEN excluded.avatar ELSE member_devices.avatar END,
			ipv4 = excluded.ipv4,
			ipv6 = excluded.ipv6,
			mac_address = excluded.mac_address,
			is_active = 1
	`
	now := time.Now().UTC()
	d.PairedAt = now
	d.IsActive = true

	_, err := r.db.ExecContext(ctx, query,
		d.ID, d.GroupID, d.DeviceUUID, d.DeviceName, d.Avatar, d.IPv4, d.IPv6, d.MACAddress, d.BatteryLevel, d.PairedAt, now,
	)
	if err != nil {
		return fmt.Errorf("failed to upsert member device: %w", err)
	}
	return nil
}

func (r *GroupRepository) GetMemberDevices(ctx context.Context, groupID string) ([]domain.MemberDevice, error) {
	query := `
		SELECT id, group_id, device_uuid, device_name, avatar, ipv4, ipv6, mac_address, battery_level, is_active, paired_at, last_seen_at
		FROM member_devices
		WHERE group_id = ?
		ORDER BY last_seen_at DESC, paired_at DESC
	`
	rows, err := r.db.QueryContext(ctx, query, groupID)
	if err != nil {
		return nil, fmt.Errorf("query member devices failed: %w", err)
	}
	defer rows.Close()

	var devices []domain.MemberDevice
	for rows.Next() {
		var d domain.MemberDevice
		var avatar, ipv4, ipv6, mac sql.NullString
		var lastSeen sql.NullTime
		var isActiveInt int

		err := rows.Scan(
			&d.ID, &d.GroupID, &d.DeviceUUID, &d.DeviceName, &avatar,
			&ipv4, &ipv6, &mac, &d.BatteryLevel, &isActiveInt, &d.PairedAt, &lastSeen,
		)
		if err != nil {
			return nil, fmt.Errorf("scan member device failed: %w", err)
		}

		d.IsActive = (isActiveInt == 1)
		if avatar.Valid {
			d.Avatar = avatar.String
		}
		if ipv4.Valid {
			d.IPv4 = ipv4.String
		}
		if ipv6.Valid {
			d.IPv6 = ipv6.String
		}
		if mac.Valid {
			d.MACAddress = mac.String
		}
		if lastSeen.Valid {
			d.LastSeenAt = &lastSeen.Time
		}
		devices = append(devices, d)
	}
	return devices, nil
}

func (r *GroupRepository) GetDeviceByID(ctx context.Context, deviceID string) (*domain.MemberDevice, error) {
	query := `
		SELECT id, group_id, device_uuid, device_name, avatar, ipv4, ipv6, mac_address, battery_level, is_active, paired_at, last_seen_at
		FROM member_devices
		WHERE id = ?
	`
	return r.scanDevice(r.db.QueryRowContext(ctx, query, deviceID))
}

func (r *GroupRepository) GetDeviceByUUID(ctx context.Context, groupID, deviceUUID string) (*domain.MemberDevice, error) {
	query := `
		SELECT id, group_id, device_uuid, device_name, avatar, ipv4, ipv6, mac_address, battery_level, is_active, paired_at, last_seen_at
		FROM member_devices
		WHERE group_id = ? AND device_uuid = ?
	`
	return r.scanDevice(r.db.QueryRowContext(ctx, query, groupID, deviceUUID))
}

func (r *GroupRepository) DeleteMemberDevice(ctx context.Context, groupID, deviceID string) error {
	query := `DELETE FROM member_devices WHERE group_id = ? AND id = ?`
	_, err := r.db.ExecContext(ctx, query, groupID, deviceID)
	if err != nil {
		return fmt.Errorf("failed to delete member device: %w", err)
	}
	return nil
}

func (r *GroupRepository) UpdateDeviceHeartbeat(ctx context.Context, deviceID string, battery int) error {
	query := `UPDATE member_devices SET last_seen_at = ?, battery_level = ?, is_active = 1 WHERE id = ?`
	now := time.Now().UTC()
	_, err := r.db.ExecContext(ctx, query, now, battery, deviceID)
	return err
}

func (r *GroupRepository) UpdateDeviceAvatar(ctx context.Context, deviceID, avatar string) error {
	query := `UPDATE member_devices SET avatar = ? WHERE id = ?`
	_, err := r.db.ExecContext(ctx, query, avatar, deviceID)
	if err != nil {
		return fmt.Errorf("failed to update device avatar: %w", err)
	}
	return nil
}

func (r *GroupRepository) scanDevice(row *sql.Row) (*domain.MemberDevice, error) {
	var d domain.MemberDevice
	var avatar, ipv4, ipv6, mac sql.NullString
	var lastSeen sql.NullTime
	var isActiveInt int

	err := row.Scan(
		&d.ID, &d.GroupID, &d.DeviceUUID, &d.DeviceName, &avatar,
		&ipv4, &ipv6, &mac, &d.BatteryLevel, &isActiveInt, &d.PairedAt, &lastSeen,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, fmt.Errorf("scan device failed: %w", err)
	}

	d.IsActive = (isActiveInt == 1)
	if avatar.Valid {
		d.Avatar = avatar.String
	}
	if ipv4.Valid {
		d.IPv4 = ipv4.String
	}
	if ipv6.Valid {
		d.IPv6 = ipv6.String
	}
	if mac.Valid {
		d.MACAddress = mac.String
	}
	if lastSeen.Valid {
		d.LastSeenAt = &lastSeen.Time
	}
	return &d, nil
}
