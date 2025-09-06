package core

import "context"

// UserRepository defines operations for user management.
type UserRepository interface {
	// CreateUser adds a new user to the database.
	CreateUser(ctx context.Context, username, email, hashedPassword string) error
	// GetUserByEmail retrieves a user by their email address.
	GetUserByEmail(ctx context.Context, email string) (*User, error)
	// SearchUsers finds users matching a query string.
	SearchUsers(ctx context.Context, query string) ([]*User, error)
}

// ChannelRepository defines operations for channel management.
type ChannelRepository interface {
	// CreateChannel adds a new channel to the database.
	CreateChannel(ctx context.Context, name, channelType string) (*Channel, error)
	// GetChannels retrieves all non-DM channels.
	GetChannels(ctx context.Context) ([]*Channel, error)
	// GetOrCreateDM finds or creates a direct message channel between two users.
	GetOrCreateDM(ctx context.Context, user1ID, user2ID int) (*Channel, error)
	// GetUserDMs retrieves all direct message channels for a specific user.
	GetUserDMs(ctx context.Context, userID int) ([]*Channel, error)
	// UpdateUserChannelReadStatus updates the last read message ID for a user in a channel.
	UpdateUserChannelReadStatus(ctx context.Context, userID, channelID, lastReadMessageID int) error
}

// MessageRepository defines operations for message management.
type MessageRepository interface {
	// SaveMessage stores a new message in the database.
	SaveMessage(ctx context.Context, channelID, userID int, content string) error
	// GetMessages retrieves a list of messages for a specific channel.
	GetMessages(ctx context.Context, channelID int, limit int) ([]*Message, error)
	// GetLatestMessageID retrieves the ID of the most recent message in a channel.
	GetLatestMessageID(ctx context.Context, channelID int) (int, error)
}

// Repository is a composite of all repository interfaces.
// This can be useful for wiring up dependencies in main.go, while services
// should depend on the smaller, more specific interfaces.
type Repository interface {
	UserRepository
	ChannelRepository
	MessageRepository

	// Close terminates the database connection.
	Close() error
}
