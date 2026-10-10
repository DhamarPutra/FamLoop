package domain

import "context"

type UserRepository interface {
	Create(ctx context.Context, user *User) error
	GetByID(ctx context.Context, id string) (*User, error)
	GetByEmail(ctx context.Context, email string) (*User, error)
	GetByUsername(ctx context.Context, username string) (*User, error)
	GetByPhone(ctx context.Context, phone string) (*User, error)
	UpdateAvatar(ctx context.Context, id, avatar string) error
	UpdateProfile(ctx context.Context, user *User) error
}

type GroupRepository interface {
	Create(ctx context.Context, group *FamilyGroup) error
	GetByID(ctx context.Context, id string) (*FamilyGroup, error)
	GetByParentID(ctx context.Context, parentID string) ([]FamilyGroup, error)
	DeleteGroup(ctx context.Context, groupID string) error
	AddMemberDevice(ctx context.Context, device *MemberDevice) error
	GetMemberDevices(ctx context.Context, groupID string) ([]MemberDevice, error)
	GetDeviceByID(ctx context.Context, deviceID string) (*MemberDevice, error)
	GetDeviceByUUID(ctx context.Context, groupID, deviceUUID string) (*MemberDevice, error)
	DeleteMemberDevice(ctx context.Context, groupID, deviceID string) error
	UpdateDeviceHeartbeat(ctx context.Context, deviceID string, battery int) error
	UpdateDeviceAvatar(ctx context.Context, deviceID, avatar string) error
}

type LocationRepository interface {
	SavePoint(ctx context.Context, point *LocationPoint) error
	SaveBatch(ctx context.Context, points []LocationPoint) (int, error)
	GetLatestGroupLocations(ctx context.Context, groupID string) ([]MemberLatestLocation, error)
	GetDeviceHistory(ctx context.Context, deviceID string, limit int) ([]LocationPoint, error)
}
