package handlers

// WSMessage represents a WebSocket message exchanged between client and server.
type WSMessage struct {
	Type      string `json:"type"`
	Content   string `json:"content,omitempty"`
	ChannelID int    `json:"channel_id,omitempty"`
	Username  string `json:"username,omitempty"`
	UserID    int    `json:"user_id,omitempty"`
	Timestamp string `json:"timestamp,omitempty"`
}

// LoginRequest defines the shape of the login request body.
type LoginRequest struct {
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required"`
}

// SignupRequest defines the shape of the signup request body.
type SignupRequest struct {
	Username string `json:"username" binding:"required,min=3,max=20"`
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required,min=6"`
}

// CreateChannelRequest defines the shape of the create channel request body.
type CreateChannelRequest struct {
	Name string `json:"name" binding:"required,min=1,max=50"`
	Type string `json:"type" binding:"required,oneof=general dm"`
}

// SearchRequest defines the shape of the search request body.
type SearchRequest struct {
	Query string `json:"query" binding:"required"`
}
