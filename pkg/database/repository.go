package database

import (
	"context"
	"github.com/divyeshmangla/nexus/internal/models"
)

type Repository interface {
	// Users
	CreateUser(ctx context.Context, username, email, hashedPassword string) error
	GetUserByEmail(ctx context.Context, email string) (*models.User, error)
	
	// Messages
	SaveMessage(ctx context.Context, channelID, userID int, content, username string) error
	GetRecentMessages(ctx context.Context, channelID int, limit int) ([]*models.Message, error)
	
	// Servers & Channels
	GetUserServers(ctx context.Context, userID int) ([]*models.Server, error)
	GetServerChannels(ctx context.Context, serverID int) ([]*models.Channel, error)
	CreateServer(ctx context.Context, name string, ownerID int) (*models.Server, error)
	JoinServer(ctx context.Context, serverID, userID int) error
	
	// DMs
	GetOrCreateDM(ctx context.Context, user1ID, user2ID int) (*models.Channel, error)
	GetUserDMs(ctx context.Context, userID int) ([]*models.Channel, error)
	SearchUsers(ctx context.Context, query string) ([]*models.User, error)
	
	// Read Status
	MarkChannelRead(ctx context.Context, userID, channelID int) error
	GetUnreadChannels(ctx context.Context, userID int) ([]int, error)
	
	Close() error
}