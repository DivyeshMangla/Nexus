package websocket

import (
	"context"
	"encoding/json"
	"log"
	"time"

	"github.com/divyeshmangla/nexus/internal/handlers"
	"github.com/gorilla/websocket"
)

const (
	writeWait  = 10 * time.Second
	pongWait   = 60 * time.Second
	pingPeriod = 54 * time.Second
)

func (c *Client) readPump() {
	defer func() {
		c.hub.unregister <- c
		c.conn.Close()
	}()

	c.conn.SetReadDeadline(time.Now().Add(pongWait))
	c.conn.SetPongHandler(func(string) error {
		c.conn.SetReadDeadline(time.Now().Add(pongWait))
		return nil
	})

	for {
		var msg handlers.WSMessage
		err := c.conn.ReadJSON(&msg)
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				log.Printf("WebSocket error: %v", err)
			}
			break
		}

		if msg.Type == "message" && msg.Content != "" {
			// Use channel ID from message, default to general if not set
			channelID := msg.ChannelID
			if channelID == 0 {
				channelID = 1 // General channel
			}

			// Save message to database
			ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
			err := c.hub.messageService.SaveMessage(ctx, channelID, c.userID, msg.Content)
			cancel()

			if err != nil {
				log.Printf("Failed to save message: %v", err)
				continue
			}

			// Broadcast message with timestamp
			msg.UserID = c.userID
			msg.Username = c.username
			msg.ChannelID = channelID
			msg.Timestamp = time.Now().UTC().Format(time.RFC3339)
			data, _ := json.Marshal(msg)
			c.hub.broadcast <- data
		} else if msg.Type == "switch_channel" {
			// Send recent messages for the channel
			c.sendRecentMessages(msg.ChannelID)
		}
	}
}

func (c *Client) sendRecentMessages(channelID int) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	messages, err := c.hub.messageService.GetMessages(ctx, channelID)
	if err != nil {
		log.Printf("Failed to get recent messages: %v", err)
		return
	}

	for _, msg := range messages {
		wsMsg := handlers.WSMessage{
			Type:      "message",
			Content:   msg.Content,
			Username:  msg.Username,
			UserID:    msg.UserID,
			ChannelID: msg.ChannelID,
			Timestamp: msg.CreatedAt.UTC().Format(time.RFC3339),
		}

		data, _ := json.Marshal(wsMsg)
		select {
		case c.send <- data:
		default:
			return
		}
	}
}

func (c *Client) writePump() {
	ticker := time.NewTicker(pingPeriod)
	defer func() {
		ticker.Stop()
		c.conn.Close()
	}()

	for {
		select {
		case message, ok := <-c.send:
			c.conn.SetWriteDeadline(time.Now().Add(writeWait))
			if !ok {
				c.conn.WriteMessage(websocket.CloseMessage, []byte{})
				return
			}

			if err := c.conn.WriteMessage(websocket.TextMessage, message); err != nil {
				return
			}

		case <-ticker.C:
			c.conn.SetWriteDeadline(time.Now().Add(writeWait))
			if err := c.conn.WriteMessage(websocket.PingMessage, nil); err != nil {
				return
			}
		}
	}
}
