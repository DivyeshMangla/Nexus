package models

import "time"

type User struct {
	ID       int    `json:"id" db:"id"`
	Username string `json:"username" db:"username"`
	Email    string `json:"email" db:"email"`
	Password string `json:"-" db:"password"`
}

type Server struct {
	ID   int    `json:"id" db:"id"`
	Name string `json:"name" db:"name"`
}

type Channel struct {
	ID       int    `json:"id" db:"id"`
	ServerID int    `json:"server_id" db:"server_id"`
	Name     string `json:"name" db:"name"`
}

type Message struct {
	ID        int       `json:"id" db:"id"`
	ChannelID int       `json:"channel_id" db:"channel_id"`
	UserID    int       `json:"user_id" db:"user_id"`
	Content   string    `json:"content" db:"content"`
	CreatedAt time.Time `json:"created_at" db:"created_at"`
	Username  string    `json:"username,omitempty"`
}