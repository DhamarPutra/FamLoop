package sqlite

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"famloop/internal/domain"
)

type LocationRepository struct {
	db *sql.DB
}

func NewLocationRepository(db *sql.DB) *LocationRepository {
	return &LocationRepository{db: db}
}

func (r *LocationRepository) SavePoint(ctx context.Context, p *domain.LocationPoint) error {
	query := `
		INSERT OR IGNORE INTO location_history (
			device_id, group_id, latitude, longitude, accuracy, speed, heading, recorded_at, synced_at
		) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)
	`
	now := time.Now().UTC()
	p.SyncedAt = now

	res, err := r.db.ExecContext(ctx, query,
		p.DeviceID, p.GroupID, p.Latitude, p.Longitude, p.Accuracy, p.Speed, p.Heading, p.RecordedAt, p.SyncedAt,
	)
	if err != nil {
		return fmt.Errorf("failed to save location point: %w", err)
	}

	id, _ := res.LastInsertId()
	p.ID = id
	return nil
}

func (r *LocationRepository) SaveBatch(ctx context.Context, points []domain.LocationPoint) (int, error) {
	if len(points) == 0 {
		return 0, nil
	}

	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return 0, fmt.Errorf("failed to begin tx: %w", err)
	}
	defer tx.Rollback()

	stmt, err := tx.PrepareContext(ctx, `
		INSERT OR IGNORE INTO location_history (
			device_id, group_id, latitude, longitude, accuracy, speed, heading, recorded_at, synced_at
		) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)
	`)
	if err != nil {
		return 0, fmt.Errorf("failed to prepare batch stmt: %w", err)
	}
	defer stmt.Close()

	now := time.Now().UTC()
	insertedCount := 0

	for _, p := range points {
		res, err := stmt.ExecContext(ctx,
			p.DeviceID, p.GroupID, p.Latitude, p.Longitude, p.Accuracy, p.Speed, p.Heading, p.RecordedAt, now,
		)
		if err != nil {
			return 0, fmt.Errorf("failed to insert batch point: %w", err)
		}
		rows, _ := res.RowsAffected()
		if rows > 0 {
			insertedCount++
		}
	}

	if err := tx.Commit(); err != nil {
		return 0, fmt.Errorf("failed to commit batch tx: %w", err)
	}

	return insertedCount, nil
}

func (r *LocationRepository) GetLatestGroupLocations(ctx context.Context, groupID string) ([]domain.MemberLatestLocation, error) {
	query := `
		WITH RankedLocations AS (
			SELECT 
				l.device_id,
				l.latitude,
				l.longitude,
				l.accuracy,
				l.speed,
				l.heading,
				l.recorded_at,
				ROW_NUMBER() OVER (PARTITION BY l.device_id ORDER BY l.recorded_at DESC) as rn
			FROM location_history l
			WHERE l.group_id = ?
		)
		SELECT 
			d.id,
			d.device_uuid,
			d.device_name,
			COALESCE(d.avatar, ''),
			d.battery_level,
			d.is_active,
			COALESCE(rl.latitude, 0.0),
			COALESCE(rl.longitude, 0.0),
			COALESCE(rl.accuracy, 0.0),
			COALESCE(rl.speed, 0.0),
			COALESCE(rl.heading, 0.0),
			COALESCE(rl.recorded_at, 0),
			COALESCE(d.last_seen_at, d.paired_at)
		FROM member_devices d
		LEFT JOIN RankedLocations rl ON d.id = rl.device_id AND rl.rn = 1
		WHERE d.group_id = ?
		ORDER BY d.device_name ASC
	`

	rows, err := r.db.QueryContext(ctx, query, groupID, groupID)
	if err != nil {
		return nil, fmt.Errorf("query latest group locations failed: %w", err)
	}
	defer rows.Close()

	var locations []domain.MemberLatestLocation
	for rows.Next() {
		var loc domain.MemberLatestLocation
		var isActiveInt int
		var lastSeenRaw any
		err := rows.Scan(
			&loc.DeviceID, &loc.DeviceUUID, &loc.DeviceName, &loc.Avatar,
			&loc.BatteryLevel, &isActiveInt,
			&loc.Latitude, &loc.Longitude, &loc.Accuracy, &loc.Speed, &loc.Heading, &loc.RecordedAt,
			&lastSeenRaw,
		)
		if err != nil {
			return nil, fmt.Errorf("scan latest location failed: %w", err)
		}
		loc.IsActive = (isActiveInt == 1)
		loc.LastSeenAt = parseSQLiteTime(lastSeenRaw)
		locations = append(locations, loc)
	}
	return locations, nil
}

func parseSQLiteTime(val any) time.Time {
	if val == nil {
		return time.Time{}
	}
	switch v := val.(type) {
	case time.Time:
		return v
	case string:
		formats := []string{
			"2006-01-02 15:04:05",
			"2006-01-02 15:04:05.999999999",
			time.RFC3339,
			"2006-01-02T15:04:05Z",
		}
		for _, f := range formats {
			if t, err := time.Parse(f, v); err == nil {
				return t
			}
		}
	}
	return time.Time{}
}

func (r *LocationRepository) GetDeviceHistory(ctx context.Context, deviceID string, limit int) ([]domain.LocationPoint, error) {
	if limit <= 0 || limit > 1000 {
		limit = 100
	}

	query := `
		SELECT id, device_id, group_id, latitude, longitude, accuracy, speed, heading, recorded_at, synced_at
		FROM location_history
		WHERE device_id = ?
		ORDER BY recorded_at DESC
		LIMIT ?
	`

	rows, err := r.db.QueryContext(ctx, query, deviceID, limit)
	if err != nil {
		return nil, fmt.Errorf("query device history failed: %w", err)
	}
	defer rows.Close()

	var points []domain.LocationPoint
	for rows.Next() {
		var p domain.LocationPoint
		err := rows.Scan(
			&p.ID, &p.DeviceID, &p.GroupID, &p.Latitude, &p.Longitude,
			&p.Accuracy, &p.Speed, &p.Heading, &p.RecordedAt, &p.SyncedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("scan history point failed: %w", err)
		}
		points = append(points, p)
	}
	return points, nil
}
