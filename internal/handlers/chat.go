package handlers

import (
	"context"
	"log"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/divyeshmangla/nexus/internal/core"
)

type ChatHandler struct {
	service *core.Service
}

func NewChatHandler(service *core.Service) *ChatHandler {
	return &ChatHandler{service: service}
}

func (h *ChatHandler) GetChannels(c *gin.Context) {
	ctx, cancel := context.WithTimeout(c.Request.Context(), 5*time.Second)
	defer cancel()

	channels, err := h.service.GetChannels(ctx)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get channels"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"channels": channels})
}

func (h *ChatHandler) CreateChannel(c *gin.Context) {
	var req core.CreateChannelRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request"})
		return
	}

	ctx, cancel := context.WithTimeout(c.Request.Context(), 5*time.Second)
	defer cancel()

	channel, err := h.service.CreateChannel(ctx, req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create channel"})
		return
	}

	c.JSON(http.StatusCreated, gin.H{"channel": channel})
}

func (h *ChatHandler) GetMessages(c *gin.Context) {
	channelID, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid channel ID"})
		return
	}

	ctx, cancel := context.WithTimeout(c.Request.Context(), 5*time.Second)
	defer cancel()

	messages, err := h.service.GetMessages(ctx, channelID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get messages"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"messages": messages})
}

func (h *ChatHandler) SearchUsers(c *gin.Context) {
	query := c.Query("q")
	if query == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Query required"})
		return
	}

	ctx, cancel := context.WithTimeout(c.Request.Context(), 5*time.Second)
	defer cancel()

	users, err := h.service.SearchUsers(ctx, query)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Search failed"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"users": users})
}

func (h *ChatHandler) CreateDM(c *gin.Context) {
	userID, exists := c.Get("userID")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}

	var req struct {
		UserID int `json:"user_id" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request"})
		return
	}

	ctx, cancel := context.WithTimeout(c.Request.Context(), 5*time.Second)
	defer cancel()

	channel, err := h.service.GetOrCreateDM(ctx, userID.(int), req.UserID)
	if err != nil {
		log.Printf("Failed to create DM: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create DM"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"id": channel.ID, "name": channel.Name, "type": channel.Type})
}

func (h *ChatHandler) GetDMs(c *gin.Context) {
	userID, exists := c.Get("userID")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}

	ctx, cancel := context.WithTimeout(c.Request.Context(), 5*time.Second)
	defer cancel()

	dms, err := h.service.GetUserDMs(ctx, userID.(int))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get DMs"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"dms": dms})
}

func (h *ChatHandler) MarkChannelAsRead(c *gin.Context) {
	userID, exists := c.Get("userID")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}

	channelID, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid channel ID"})
		return
	}

	ctx, cancel := context.WithTimeout(c.Request.Context(), 5*time.Second)
	defer cancel()

	latestMessageID, err := h.service.GetLatestMessageID(ctx, channelID)
	if err != nil {
		log.Printf("Failed to get latest message ID for channel %d: %v", channelID, err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to mark channel as read"})
		return
	}

	err = h.service.UpdateUserChannelReadStatus(ctx, userID.(int), channelID, latestMessageID)
	if err != nil {
		log.Printf("Failed to update read status for user %d, channel %d: %v", userID.(int), channelID, err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to mark channel as read"})
		return
	}

	c.Status(http.StatusOK)
}