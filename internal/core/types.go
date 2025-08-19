package core

import "time"

// User represents a system user
type User struct {
	ID       int    `json:"id" db:"id"`
	Username string `json:"username" db:"username"`
	Email    string `json:"email" db:"email"`
	Password string `json:"-" db:"password"`
}

// Channel represents a communication channel
type Channel struct {
	ID   int    `json:"id" db:"id"`
	Name string `json:"name" db:"name"`
	Type string `json:"type" db:"type"` // "general" or "dm"
}

// Message represents a chat message
type Message struct {
	ID        int       `json:"id" db:"id"`
	ChannelID int       `json:"channel_id" db:"channel_id"`
	UserID    int       `json:"user_id" db:"user_id"`
	Content   string    `json:"content" db:"content"`
	CreatedAt time.Time `json:"created_at" db:"created_at"`
	Username  string    `json:"username,omitempty"`
}

// WSMessage represents WebSocket message
type WSMessage struct {
	Type      string `json:"type"`
	Content   string `json:"content,omitempty"`
	ChannelID int    `json:"channel_id,omitempty"`
	Username  string `json:"username,omitempty"`
	UserID    int    `json:"user_id,omitempty"`
	Timestamp string `json:"timestamp,omitempty"`
}

// Request types
type LoginRequest struct {
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required"`
}

type SignupRequest struct {
	Username string `json:"username" binding:"required,min=3,max=20"`
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required,min=6"`
}

type CreateChannelRequest struct {
	Name string `json:"name" binding:"required,min=1,max=50"`
	Type string `json:"type" binding:"required,oneof=general dm"`
}