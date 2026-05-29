package model

import "time"

type Group struct {
	ID         string    `json:"id"`
	Name       string    `json:"name"`
	InviteCode string    `json:"invite_code"`
	CreatedAt  time.Time `json:"created_at"`
}

type Member struct {
	UserID      string    `json:"user_id"`
	DisplayName string    `json:"display_name"`
	AvatarURL   *string   `json:"avatar_url,omitempty"`
	JoinedAt    time.Time `json:"joined_at"`
}

type GroupWithMembers struct {
	Group
	Members []Member `json:"members"`
}
