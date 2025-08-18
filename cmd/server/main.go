package main

import (
	"log"
	"net/http"
	
	"github.com/gin-gonic/gin"
	"github.com/divyeshmangla/nexus/config"
	"github.com/divyeshmangla/nexus/pkg/database"
	"github.com/divyeshmangla/nexus/internal/websocket"
)

func main() {
	cfg := config.Load()
	
	db, err := database.Connect(cfg.DBHost, cfg.DBPort, cfg.DBUser, cfg.DBPassword, cfg.DBName)
	if err != nil {
		log.Fatal("Failed to connect to database:", err)
	}
	defer db.Close()
	
	hub := websocket.NewHub()
	go hub.Run()
	
	r := gin.Default()
	r.LoadHTMLGlob("web/templates/*")
	r.Static("/static", "./web/static")
	
	r.GET("/", func(c *gin.Context) {
		c.HTML(http.StatusOK, "index.html", nil)
	})
	
	r.GET("/ws", func(c *gin.Context) {
		c.String(200, "WebSocket endpoint - needs implementation")
	})
	
	log.Printf("Server starting on port %s", cfg.Port)
	log.Fatal(http.ListenAndServe(":"+cfg.Port, r))
}