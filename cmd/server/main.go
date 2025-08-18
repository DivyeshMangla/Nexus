package main

import (
	"log"
	"net/http"
	
	"github.com/gin-gonic/gin"
	"github.com/divyeshmangla/nexus/config"
	"github.com/divyeshmangla/nexus/pkg/database"
	"github.com/divyeshmangla/nexus/internal/websocket"
	"github.com/divyeshmangla/nexus/internal/user"
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
	
	userHandler := user.NewHandler(db, cfg.JWTSecret)
	
	api := r.Group("/api")
	{
		api.POST("/signup", userHandler.Signup)
		api.POST("/login", userHandler.Login)
	}
	
	r.GET("/", func(c *gin.Context) {
		c.Redirect(http.StatusFound, "/login")
	})
	
	r.GET("/login", func(c *gin.Context) {
		c.HTML(http.StatusOK, "login.html", nil)
	})
	
	r.GET("/signup", func(c *gin.Context) {
		c.HTML(http.StatusOK, "signup.html", nil)
	})
	
	r.GET("/chat", func(c *gin.Context) {
		c.HTML(http.StatusOK, "index.html", nil)
	})
	
	r.GET("/ws", hub.HandleWebSocket(cfg.JWTSecret))
	
	log.Printf("Server starting on port %s", cfg.Port)
	log.Fatal(http.ListenAndServe(":"+cfg.Port, r))
}