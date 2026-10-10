package domain

import "time"

type FamilyGroup struct {
	ID         string    `json:"id"`
	Name       string    `json:"name"`
	ParentID   string    `json:"parent_id"`
	InviteCode string    `json:"invite_code,omitempty"`
	CreatedAt  time.Time `json:"created_at"`
}
