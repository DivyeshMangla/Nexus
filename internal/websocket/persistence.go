package websocket

import (
	"context"
	"log"
	"time"
	"github.com/divyeshmangla/nexus/internal/core"
)

func (h *Hub) saveMessage(msg core.WSMessage) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// Use channel ID from message, default to general if not set
	channelID := msg.ChannelID
	if channelID == 0 {
		channelID = 1 // General channel
	}
	
	err := h.service.SaveMessage(ctx, channelID, msg.UserID, msg.Content)
	if err != nil {
		log.Printf("Failed to save message: %v", err)
	}
}