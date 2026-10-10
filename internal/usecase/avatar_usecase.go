package usecase

import (
	"context"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"time"

	"famloop/internal/domain"
)

var (
	ErrInvalidImageType = errors.New("format gambar tidak didukung (gunakan JPG, PNG, atau WebP)")
	ErrFileTooLarge     = errors.New("ukuran gambar melebihi batas 5MB")
)

type AvatarUseCase struct {
	userRepo    domain.UserRepository
	groupRepo   domain.GroupRepository
	broadcaster domain.EventBroadcaster
}

func NewAvatarUseCase(
	userRepo domain.UserRepository,
	groupRepo domain.GroupRepository,
	broadcaster domain.EventBroadcaster,
) *AvatarUseCase {
	return &AvatarUseCase{
		userRepo:    userRepo,
		groupRepo:   groupRepo,
		broadcaster: broadcaster,
	}
}

func (uc *AvatarUseCase) UploadUserAvatar(
	ctx context.Context,
	userID string,
	fileReader io.Reader,
	originalName string,
	uploadBaseDir string,
) (string, error) {
	ext := strings.ToLower(filepath.Ext(originalName))
	if !isAllowedImageExt(ext) {
		return "", ErrInvalidImageType
	}

	avatarsDir := filepath.Join(uploadBaseDir, "avatars")
	if err := os.MkdirAll(avatarsDir, 0755); err != nil {
		return "", fmt.Errorf("failed to create avatars directory: %w", err)
	}

	filename := fmt.Sprintf("user_%s_%d%s", userID, time.Now().Unix(), ext)
	dstPath := filepath.Join(avatarsDir, filename)

	dst, err := os.Create(dstPath)
	if err != nil {
		return "", fmt.Errorf("failed to save avatar file: %w", err)
	}
	defer dst.Close()

	written, err := io.Copy(dst, fileReader)
	if err != nil {
		return "", fmt.Errorf("failed to write avatar file: %w", err)
	}
	if written > 5*1024*1024 { // 5MB limit
		_ = os.Remove(dstPath)
		return "", ErrFileTooLarge
	}

	avatarURL := "/uploads/avatars/" + filename
	if err := uc.userRepo.UpdateAvatar(ctx, userID, avatarURL); err != nil {
		return "", fmt.Errorf("failed to update user avatar in db: %w", err)
	}

	return avatarURL, nil
}

func (uc *AvatarUseCase) UploadDeviceAvatar(
	ctx context.Context,
	deviceID string,
	groupID string,
	fileReader io.Reader,
	originalName string,
	uploadBaseDir string,
) (string, error) {
	ext := strings.ToLower(filepath.Ext(originalName))
	if !isAllowedImageExt(ext) {
		return "", ErrInvalidImageType
	}

	avatarsDir := filepath.Join(uploadBaseDir, "avatars")
	if err := os.MkdirAll(avatarsDir, 0755); err != nil {
		return "", fmt.Errorf("failed to create avatars directory: %w", err)
	}

	filename := fmt.Sprintf("dev_%s_%d%s", deviceID, time.Now().Unix(), ext)
	dstPath := filepath.Join(avatarsDir, filename)

	dst, err := os.Create(dstPath)
	if err != nil {
		return "", fmt.Errorf("failed to save device avatar file: %w", err)
	}
	defer dst.Close()

	written, err := io.Copy(dst, fileReader)
	if err != nil {
		return "", fmt.Errorf("failed to write device avatar file: %w", err)
	}
	if written > 5*1024*1024 {
		_ = os.Remove(dstPath)
		return "", ErrFileTooLarge
	}

	avatarURL := "/uploads/avatars/" + filename
	if err := uc.groupRepo.UpdateDeviceAvatar(ctx, deviceID, avatarURL); err != nil {
		return "", fmt.Errorf("failed to update device avatar in db: %w", err)
	}

	// Siarkan event pembaruan avatar ke seluruh koneksi SSE grup
	if uc.broadcaster != nil {
		uc.broadcaster.BroadcastToGroup(groupID, "MemberAvatarUpdated", map[string]any{
			"device_id":  deviceID,
			"group_id":   groupID,
			"avatar_url": avatarURL,
		})
	}

	return avatarURL, nil
}

func isAllowedImageExt(ext string) bool {
	switch ext {
	case ".jpg", ".jpeg", ".png", ".webp":
		return true
	default:
		return false
	}
}
