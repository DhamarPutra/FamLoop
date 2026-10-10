package usecase

import (
	"context"
	"crypto/rand"
	"errors"
	"fmt"
	"math/big"
	"strings"

	"famloop/internal/domain"

	"github.com/google/uuid"
)

var (
	ErrGroupNotFound     = errors.New("grup keluarga tidak ditemukan")
	ErrGroupUnauthorized = errors.New("anda tidak memiliki izin untuk mengelola grup ini")
	ErrInvalidQRData     = errors.New("data QR code perangkat tidak valid")
)

type QRScanPayload struct {
	DeviceUUID string `json:"device_uuid"`
	DeviceName string `json:"device_name"`
	Avatar     string `json:"avatar,omitempty"`
	IPv4       string `json:"ipv4,omitempty"`
	IPv6       string `json:"ipv6,omitempty"`
	MACAddress string `json:"mac_address,omitempty"`
}

type GroupUseCase struct {
	groupRepo   domain.GroupRepository
	broadcaster domain.EventBroadcaster
}

func NewGroupUseCase(groupRepo domain.GroupRepository, broadcaster domain.EventBroadcaster) *GroupUseCase {
	return &GroupUseCase{
		groupRepo:   groupRepo,
		broadcaster: broadcaster,
	}
}

func (uc *GroupUseCase) CreateGroup(ctx context.Context, parentID, name string) (*domain.FamilyGroup, error) {
	name = strings.TrimSpace(name)
	if len(name) < 2 {
		return nil, errors.New("nama grup minimal 2 karakter")
	}

	inviteCode := generateInviteCode()
	group := &domain.FamilyGroup{
		ID:         uuid.New().String(),
		Name:       name,
		ParentID:   parentID,
		InviteCode: inviteCode,
	}

	if err := uc.groupRepo.Create(ctx, group); err != nil {
		return nil, fmt.Errorf("failed to create group: %w", err)
	}

	return group, nil
}

func (uc *GroupUseCase) GetParentGroups(ctx context.Context, parentID string) ([]domain.FamilyGroup, error) {
	return uc.groupRepo.GetByParentID(ctx, parentID)
}

func (uc *GroupUseCase) GetGroupDetails(ctx context.Context, parentID, groupID string) (*domain.FamilyGroup, []domain.MemberDevice, error) {
	group, err := uc.groupRepo.GetByID(ctx, groupID)
	if err != nil {
		return nil, nil, err
	}
	if group == nil {
		return nil, nil, ErrGroupNotFound
	}

	if parentID != "" && group.ParentID != parentID {
		return nil, nil, ErrGroupUnauthorized
	}

	devices, err := uc.groupRepo.GetMemberDevices(ctx, groupID)
	if err != nil {
		return nil, nil, err
	}

	return group, devices, nil
}

func (uc *GroupUseCase) PairDeviceFromQR(ctx context.Context, parentID, groupID string, payload QRScanPayload) (*domain.MemberDevice, error) {
	group, err := uc.groupRepo.GetByID(ctx, groupID)
	if err != nil {
		return nil, err
	}
	if group == nil {
		return nil, ErrGroupNotFound
	}
	if group.ParentID != parentID {
		return nil, ErrGroupUnauthorized
	}

	payload.DeviceUUID = strings.TrimSpace(payload.DeviceUUID)
	payload.DeviceName = strings.TrimSpace(payload.DeviceName)

	if payload.DeviceUUID == "" || payload.DeviceName == "" {
		return nil, ErrInvalidQRData
	}

	avatar := payload.Avatar
	if avatar == "" {
		avatar = domain.GetRandomDefaultAvatar()
	}

	// Cek apakah device sudah pernah dipasangkan sebelumnya di grup ini
	existing, err := uc.groupRepo.GetDeviceByUUID(ctx, groupID, payload.DeviceUUID)
	if err != nil {
		return nil, err
	}

	var deviceID string
	if existing != nil {
		deviceID = existing.ID
	} else {
		deviceID = uuid.New().String()
	}

	device := &domain.MemberDevice{
		ID:           deviceID,
		GroupID:      groupID,
		DeviceUUID:   payload.DeviceUUID,
		DeviceName:   payload.DeviceName,
		Avatar:       avatar,
		IPv4:         payload.IPv4,
		IPv6:         payload.IPv6,
		MACAddress:   payload.MACAddress,
		BatteryLevel: 100,
		IsActive:     true,
	}

	if err := uc.groupRepo.AddMemberDevice(ctx, device); err != nil {
		return nil, fmt.Errorf("failed to pair member device: %w", err)
	}

	// Siarkan event ke SSE Hub agar seluruh hp keluarga menerima notifikasi anggota baru/aktif
	if uc.broadcaster != nil {
		uc.broadcaster.BroadcastToGroup(groupID, "MemberJoined", device)
	}

	return device, nil
}

func (uc *GroupUseCase) GetDevice(ctx context.Context, deviceID string) (*domain.MemberDevice, error) {
	return uc.groupRepo.GetDeviceByID(ctx, deviceID)
}

func (uc *GroupUseCase) DeleteGroup(ctx context.Context, parentID, groupID string) error {
	group, err := uc.groupRepo.GetByID(ctx, groupID)
	if err != nil {
		return err
	}
	if group == nil {
		return ErrGroupNotFound
	}
	if group.ParentID != parentID {
		return ErrGroupUnauthorized
	}

	if err := uc.groupRepo.DeleteGroup(ctx, groupID); err != nil {
		return fmt.Errorf("failed to delete group: %w", err)
	}

	if uc.broadcaster != nil {
		uc.broadcaster.BroadcastToGroup(groupID, "GroupDeleted", map[string]string{"group_id": groupID})
	}

	return nil
}

func (uc *GroupUseCase) RemoveMemberDevice(ctx context.Context, parentID, groupID, deviceID string) error {
	group, err := uc.groupRepo.GetByID(ctx, groupID)
	if err != nil {
		return err
	}
	if group == nil {
		return ErrGroupNotFound
	}
	if group.ParentID != parentID {
		return ErrGroupUnauthorized
	}

	if err := uc.groupRepo.DeleteMemberDevice(ctx, groupID, deviceID); err != nil {
		return fmt.Errorf("failed to remove member device: %w", err)
	}

	if uc.broadcaster != nil {
		uc.broadcaster.BroadcastToGroup(groupID, "MemberRemoved", map[string]string{
			"group_id":  groupID,
			"device_id": deviceID,
		})
	}

	return nil
}

func generateInviteCode() string {
	const chars = "ABCDEFGHJKLMNPQRSTUVWXYZ23456789"
	b := make([]byte, 6)
	for i := range b {
		idx, _ := rand.Int(rand.Reader, big.NewInt(int64(len(chars))))
		b[i] = chars[idx.Int64()]
	}
	return "FAM-" + string(b)
}
