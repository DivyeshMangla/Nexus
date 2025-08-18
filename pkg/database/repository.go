package database

import (
	"context"
	"github.com/divyeshmangla/nexus/internal/models"
)

type Repository interface {
	CreateUser(ctx context.Context, username, email, hashedPassword string) error
	GetUserByEmail(ctx context.Context, email string) (*models.User, error)
	SaveMessage(ctx context.Context, channelID, userID int, content, username string) error
	GetRecentMessages(ctx context.Context, channelID int, limit int) ([]*models.Message, error)
	Close() error
}