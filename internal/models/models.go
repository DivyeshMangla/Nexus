package models

import "time"

const (
	ChannelTypeText = "text"
	ChannelTypeDM   = "dm"
	GeneralChannelID = 1
)

type Server struct {
	ID        int       `json:"id" db:"id"`
	Name      string    `json:"name" db:"name"`
	OwnerID   int       `json:"owner_id" db:"owner_id"`
	CreatedAt time.Time `json:"created_at" db:"created_at"`
}

type Channel struct {
	ID        int       `json:"id" db:"id"`
	ServerID  *int      `json:"server_id" db:"server_id"`
	Name      string    `json:"name" db:"name"`
	Type      string    `json:"type" db:"type"`
	CreatedAt time.Time `json:"created_at" db:"created_at"`
}

type Message struct {
	ID        int       `json:"id" db:"id"`
	ChannelID int       `json:"channel_id" db:"channel_id"`
	UserID    int       `json:"user_id" db:"user_id"`
	Content   string    `json:"content" db:"content"`
	CreatedAt time.Time `json:"created_at" db:"created_at"`
	Username  string    `json:"username,omitempty"`
}