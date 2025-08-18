package user

import (
	"context"
	"database/sql"
	"net/http"
	"time"
	"github.com/gin-gonic/gin"
	"github.com/divyeshmangla/nexus/internal/auth"
	"github.com/divyeshmangla/nexus/internal/models"
	"github.com/divyeshmangla/nexus/pkg/database"
	"github.com/divyeshmangla/nexus/pkg/errors"
)

type Handler struct {
	db     database.Repository
	secret string
}

func NewHandler(db database.Repository, secret string) *Handler {
	return &Handler{db: db, secret: secret}
}

func (h *Handler) Signup(c *gin.Context) {
	var req models.SignupRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request data"})
		return
	}

	ctx, cancel := context.WithTimeout(c.Request.Context(), 5*time.Second)
	defer cancel()

	hashedPassword, err := auth.HashPassword(req.Password)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": errors.ErrPasswordHashing.Error()})
		return
	}

	if err := h.db.CreateUser(ctx, req.Username, req.Email, hashedPassword); err != nil {
		c.JSON(http.StatusConflict, gin.H{"error": errors.ErrUserExists.Error()})
		return
	}

	c.JSON(http.StatusCreated, gin.H{"message": "User created successfully"})
}

func (h *Handler) Login(c *gin.Context) {
	var req models.LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request data"})
		return
	}

	ctx, cancel := context.WithTimeout(c.Request.Context(), 5*time.Second)
	defer cancel()

	user, err := h.db.GetUserByEmail(ctx, req.Email)
	if err != nil {
		if err == sql.ErrNoRows {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid credentials"})
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Internal server error"})
		}
		return
	}

	if !auth.CheckPassword(req.Password, user.Password) {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid credentials"})
		return
	}

	token, err := auth.GenerateToken(user.ID, user.Username, h.secret)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": errors.ErrTokenGeneration.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"token":    token,
		"username": user.Username,
	})
}