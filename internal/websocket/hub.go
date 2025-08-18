package websocket

import (
	"encoding/json"
	"log"
	"net/http"
	"github.com/gorilla/websocket"
	"github.com/divyeshmangla/nexus/internal/models"
	"github.com/divyeshmangla/nexus/pkg/database"
)

type Hub struct {
	channels   map[int]map[*Client]bool // channelID -> clients
	broadcast  chan []byte
	register   chan *Client
	unregister chan *Client
	db         database.Repository
}

type Client struct {
	hub       *Hub
	conn      *websocket.Conn
	send      chan []byte
	userID    int
	username  string
	channelID int
}



var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		origin := r.Header.Get("Origin")
		return origin == "http://localhost:8080" || origin == "https://localhost:8080"
	},
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
}

func NewHub(db database.Repository) *Hub {
	return &Hub{
		channels:   make(map[int]map[*Client]bool),
		broadcast:  make(chan []byte),
		register:   make(chan *Client),
		unregister: make(chan *Client),
		db:         db,
	}
}

func (h *Hub) Run() {
	for {
		select {
		case client := <-h.register:
			// Default to general channel
			if client.channelID == 0 {
				client.channelID = models.GeneralChannelID
			}
			
			if h.channels[client.channelID] == nil {
				h.channels[client.channelID] = make(map[*Client]bool)
			}
			h.channels[client.channelID][client] = true
			log.Printf("Client %s joined channel %d", client.username, client.channelID)
			
			// Send recent messages to new client
			go h.sendRecentMessages(client)

		case client := <-h.unregister:
			if clients, ok := h.channels[client.channelID]; ok {
				if _, exists := clients[client]; exists {
					delete(clients, client)
					close(client.send)
					log.Printf("Client %s left channel %d", client.username, client.channelID)
					
					// Clean up empty channels
					if len(clients) == 0 {
						delete(h.channels, client.channelID)
					}
				}
			}

		case message := <-h.broadcast:
			// Save message to database
			var msg Message
			if err := json.Unmarshal(message, &msg); err == nil {
				go h.saveMessage(msg)
				
				// Broadcast to all clients (they'll filter by channel on frontend)
				for _, clients := range h.channels {
					for client := range clients {
						select {
						case client.send <- message:
						default:
							close(client.send)
							delete(clients, client)
						}
					}
				}
			}
		}
	}
}

func (h *Hub) SwitchChannel(client *Client, newChannelID int) {
	// Remove from current channel
	if clients, ok := h.channels[client.channelID]; ok {
		delete(clients, client)
		if len(clients) == 0 {
			delete(h.channels, client.channelID)
		}
	}
	
	// Add to new channel
	client.channelID = newChannelID
	if h.channels[newChannelID] == nil {
		h.channels[newChannelID] = make(map[*Client]bool)
	}
	h.channels[newChannelID][client] = true
	
	log.Printf("Client %s switched to channel %d", client.username, newChannelID)
	
	// Send recent messages from new channel
	go h.sendRecentMessages(client)
}