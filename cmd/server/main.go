package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/divyeshmangla/nexus/config"
	"github.com/divyeshmangla/nexus/internal/core"
	"github.com/divyeshmangla/nexus/internal/database"
	"github.com/divyeshmangla/nexus/internal/handlers"
	"github.com/divyeshmangla/nexus/internal/middleware"
	"github.com/divyeshmangla/nexus/internal/websocket"
)

func main() {
	cfg := config.Load()

	// Setup database
	db, err := database.Connect(cfg.DBHost, cfg.DBPort, cfg.DBUser, cfg.DBPassword, cfg.DBName)
	if err != nil {
		log.Fatal("Database connection failed:", err)
	}
	defer db.Close()

	// Setup service
	service := core.NewService(db, cfg.JWTSecret)

	// Setup handlers
	authHandler := handlers.NewAuthHandler(service)
	chatHandler := handlers.NewChatHandler(service)

	// Setup WebSocket hub
	hub := websocket.NewHub(service)
	go hub.Run()

	// Setup router
	r := gin.Default()
	r.LoadHTMLGlob("web/templates/*")
	r.Static("/static", "./web/static")

	// Routes
	api := r.Group("/api")
	{
		api.POST("/signup", authHandler.Register)
		api.POST("/login", authHandler.Login)

		protected := api.Group("/", middleware.AuthMiddleware(cfg.JWTSecret))
		{
			protected.GET("/channels", chatHandler.GetChannels)
			protected.POST("/channels", chatHandler.CreateChannel)
			protected.GET("/channels/:id/messages", chatHandler.GetMessages)
			protected.GET("/users/search", chatHandler.SearchUsers)
			protected.POST("/dms", chatHandler.CreateDM)
		}
	}

	r.GET("/", func(c *gin.Context) { c.Redirect(http.StatusFound, "/login") })
	r.GET("/login", func(c *gin.Context) { c.HTML(http.StatusOK, "login.html", nil) })
	r.GET("/signup", func(c *gin.Context) { c.HTML(http.StatusOK, "signup.html", nil) })
	r.GET("/chat", func(c *gin.Context) { c.HTML(http.StatusOK, "index.html", nil) })
	r.GET("/ws", hub.HandleWebSocket(cfg.JWTSecret))

	// Start server
	srv := &http.Server{
		Addr:    ":" + cfg.Port,
		Handler: r,
	}

	go func() {
		log.Printf("Server starting on port %s", cfg.Port)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatal("Server failed:", err)
		}
	}()

	// Graceful shutdown
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Println("Shutting down server...")
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		log.Fatal("Server shutdown failed:", err)
	}

	log.Println("Server exited")
}