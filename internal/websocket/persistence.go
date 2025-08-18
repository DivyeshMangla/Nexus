package websocket

import (
	"context"
	"encoding/json"
	"log"
	"time"
)

func (h *Hub) sendRecentMessages(client *Client) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// Get recent messages from channel 1 (default general channel)
	messages, err := h.db.GetRecentMessages(ctx, 1, 50)
	if err != nil {
		log.Printf("Failed to get recent messages: %v", err)
		return
	}

	for _, msg := range messages {
		wsMsg := Message{
			Type:      "message",
			Content:   msg.Content,
			Username:  msg.Username,
			UserID:    msg.UserID,
			Timestamp: msg.CreatedAt.Format("15:04"),
		}
		
		data, _ := json.Marshal(wsMsg)
		select {
		case client.send <- data:
		default:
			return
		}
	}
}

func (h *Hub) saveMessage(msg Message) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// Save to channel 1 (default general channel)
	err := h.db.SaveMessage(ctx, 1, msg.UserID, msg.Content, msg.Username)
	if err != nil {
		log.Printf("Failed to save message: %v", err)
	}
}