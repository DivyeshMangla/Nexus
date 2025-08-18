package websocket

import (
	"context"
	"encoding/json"
	"log"
	"time"
	"github.com/divyeshmangla/nexus/internal/models"
)

func (h *Hub) sendRecentMessages(client *Client) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// Get recent messages from client's current channel
	messages, err := h.db.GetRecentMessages(ctx, client.channelID, 50)
	if err != nil {
		log.Printf("Failed to get recent messages for channel %d: %v", client.channelID, err)
		return
	}

	for _, msg := range messages {
		wsMsg := Message{
			Type:      "message",
			Content:   msg.Content,
			Username:  msg.Username,
			UserID:    msg.UserID,
			ChannelID: msg.ChannelID,
			Timestamp: msg.CreatedAt.UTC().Format("2006-01-02T15:04:05Z"),
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

	// Use channel ID from message, default to general if not set
	channelID := msg.ChannelID
	if channelID == 0 {
		channelID = models.GeneralChannelID
	}
	
	err := h.db.SaveMessage(ctx, channelID, msg.UserID, msg.Content, msg.Username)
	if err != nil {
		log.Printf("Failed to save message: %v", err)
	}
}