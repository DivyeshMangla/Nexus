package core

import "time"

// User represents a system user.
type User struct {
	ID       int    `json:"id" db:"id"`
	Username string `json:"username" db:"username"`
	Email    string `json:"email" db:"email"`
	Password string `json:"-" db:"password"`
}

// Channel represents a communication channel.
type Channel struct {
	ID   int    `json:"id" db:"id"`
	Name string `json:"name" db:"name"`
	Type string `json:"type" db:"type"` // "general" or "dm"
}

// Message represents a chat message.
type Message struct {
	ID        int       `json:"id" db:"id"`
	ChannelID int       `json:"channel_id" db:"channel_id"`
	UserID    int       `json:"user_id" db:"user_id"`
	Content   string    `json:"content" db:"content"`
	CreatedAt time.Time `json:"created_at" db:"created_at"`
	Username  string    `json:"username,omitempty"`
}
