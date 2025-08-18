package server

import (
	"context"
	"log"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/divyeshmangla/nexus/config"
	"github.com/divyeshmangla/nexus/internal/chat"
	"github.com/divyeshmangla/nexus/internal/user"
	"github.com/divyeshmangla/nexus/internal/websocket"
	"github.com/divyeshmangla/nexus/pkg/database"
	"github.com/divyeshmangla/nexus/pkg/middleware"
)

type Server struct {
	cfg        *config.Config
	db         database.Repository
	hub        *websocket.Hub
	httpServer *http.Server
}

func New(cfg *config.Config) *Server {
	return &Server{cfg: cfg}
}

func (s *Server) Start() error {
	if err := s.setupDatabase(); err != nil {
		return err
	}

	s.setupWebSocket()
	router := s.setupRoutes()

	s.httpServer = &http.Server{
		Addr:         ":" + s.cfg.Port,
		Handler:      router,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
	}

	log.Printf("Server starting on port %s", s.cfg.Port)
	return s.httpServer.ListenAndServe()
}

func (s *Server) Shutdown(ctx context.Context) error {
	return s.httpServer.Shutdown(ctx)
}

func (s *Server) setupDatabase() error {
	db, err := database.Connect(s.cfg.DBHost, s.cfg.DBPort, s.cfg.DBUser, s.cfg.DBPassword, s.cfg.DBName)
	if err != nil {
		return err
	}
	s.db = db
	return nil
}

func (s *Server) setupWebSocket() {
	s.hub = websocket.NewHub(s.db)
	go s.hub.Run()
}

func (s *Server) setupRoutes() *gin.Engine {
	r := gin.Default()
	r.LoadHTMLGlob("web/templates/*")
	r.Static("/static", "./web/static")

	userHandler := user.NewHandler(s.db, s.cfg.JWTSecret)
	chatHandler := chat.NewHandler(s.db)

	api := r.Group("/api")
	{
		api.POST("/signup", userHandler.Signup)
		api.POST("/login", userHandler.Login)
		
		// Protected routes
		protected := api.Group("/", middleware.AuthMiddleware(s.cfg.JWTSecret))
		{
			protected.GET("/users/search", chatHandler.SearchUsers)
			protected.POST("/dms", chatHandler.CreateDM)
			protected.GET("/dms", chatHandler.GetUserDMs)
			protected.GET("/channels/:channelId/messages", chatHandler.GetChannelMessages)
			protected.POST("/channels/:channelId/read", chatHandler.MarkChannelRead)
			protected.GET("/unread", chatHandler.GetUnreadChannels)
		}
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
	r.GET("/ws", s.hub.HandleWebSocket(s.cfg.JWTSecret))

	return r
}