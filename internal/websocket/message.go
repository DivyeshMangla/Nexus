package websocket

// WebSocket message types
const (
	MessageTypeChat          = "message"
	MessageTypeSwitchChannel = "switch_channel"
	MessageTypeJoin          = "join"
	MessageTypeLeave         = "leave"
)

// WSMessage represents a WebSocket message
type WSMessage struct {
	Type      string `json:"type"`
	Content   string `json:"content,omitempty"`
	Username  string `json:"username,omitempty"`
	UserID    int    `json:"user_id,omitempty"`
	ChannelID int    `json:"channel_id,omitempty"`
	Timestamp string `json:"timestamp,omitempty"`
}

// IsChannelSwitch returns true if message is a channel switch request
func (m *WSMessage) IsChannelSwitch() bool {
	return m.Type == MessageTypeSwitchChannel
}

// IsChatMessage returns true if message is a chat message
func (m *WSMessage) IsChatMessage() bool {
	return m.Type == MessageTypeChat
}