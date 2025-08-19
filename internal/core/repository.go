package core

import "context"

// Repository defines all database operations
type Repository interface {
	// User operations
	CreateUser(ctx context.Context, username, email, hashedPassword string) error
	GetUserByEmail(ctx context.Context, email string) (*User, error)
	SearchUsers(ctx context.Context, query string) ([]*User, error)

	// Channel operations
	CreateChannel(ctx context.Context, name, channelType string) (*Channel, error)
	GetChannels(ctx context.Context) ([]*Channel, error)
	GetOrCreateDM(ctx context.Context, user1ID, user2ID int) (*Channel, error)

	// Message operations
	SaveMessage(ctx context.Context, channelID, userID int, content string) error
	GetMessages(ctx context.Context, channelID int, limit int) ([]*Message, error)
	GetLatestMessageID(ctx context.Context, channelID int) (int, error)

	// DM operations
	GetUserDMs(ctx context.Context, userID int) ([]*Channel, error)
	UpdateUserChannelReadStatus(ctx context.Context, userID, channelID, lastReadMessageID int) error

	Close() error
}