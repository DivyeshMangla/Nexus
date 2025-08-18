package websocket

type Message struct {
	Type      string `json:"type"`
	Content   string `json:"content"`
	Username  string `json:"username"`
	UserID    int    `json:"user_id"`
	Timestamp string `json:"timestamp"`
}