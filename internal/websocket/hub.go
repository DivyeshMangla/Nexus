package websocket

import (
	"encoding/json"
	"log"
	"net/http"
	"github.com/gorilla/websocket"
	"github.com/divyeshmangla/nexus/pkg/database"
)

type Hub struct {
	clients    map[*Client]bool
	broadcast  chan []byte
	register   chan *Client
	unregister chan *Client
	db         database.Repository
}

type Client struct {
	hub      *Hub
	conn     *websocket.Conn
	send     chan []byte
	userID   int
	username string
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
		clients:    make(map[*Client]bool),
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
			h.clients[client] = true
			log.Printf("Client connected: %s", client.username)
			
			// Send recent messages to new client
			go h.sendRecentMessages(client)

		case client := <-h.unregister:
			if _, ok := h.clients[client]; ok {
				delete(h.clients, client)
				close(client.send)
				log.Printf("Client disconnected: %s", client.username)
			}

		case message := <-h.broadcast:
			// Save message to database
			var msg Message
			if err := json.Unmarshal(message, &msg); err == nil {
				go h.saveMessage(msg)
			}
			
			for client := range h.clients {
				select {
				case client.send <- message:
				default:
					close(client.send)
					delete(h.clients, client)
				}
			}
		}
	}
}