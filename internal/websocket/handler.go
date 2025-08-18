package websocket

import (
	"log"
	"net/http"
	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"github.com/divyeshmangla/nexus/internal/auth"
)

func (h *Hub) HandleWebSocket(secret string) gin.HandlerFunc {
	return func(c *gin.Context) {
		token := c.Query("token")
		if token == "" {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Token required"})
			return
		}

		claims := &auth.Claims{}
		tkn, err := jwt.ParseWithClaims(token, claims, func(token *jwt.Token) (interface{}, error) {
			return []byte(secret), nil
		})

		if err != nil || !tkn.Valid {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid token"})
			return
		}

		conn, err := upgrader.Upgrade(c.Writer, c.Request, nil)
		if err != nil {
			log.Println(err)
			return
		}

		client := &Client{
			hub:      h,
			conn:     conn,
			send:     make(chan []byte, 256),
			userID:   claims.UserID,
			username: claims.Username,
		}

		client.hub.register <- client

		go client.writePump()
		go client.readPump()
	}
}