package core

import (
	"context"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"
)

// AuthService handles user authentication logic.
type AuthService struct {
	userRepo  UserRepository
	jwtSecret string
}

// NewAuthService creates a new AuthService.
func NewAuthService(userRepo UserRepository, jwtSecret string) *AuthService {
	return &AuthService{
		userRepo:  userRepo,
		jwtSecret: jwtSecret,
	}
}

// Register creates a new user.
func (s *AuthService) Register(ctx context.Context, username, email, password string) error {
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), 12)
	if err != nil {
		return err
	}
	return s.userRepo.CreateUser(ctx, username, email, string(hashedPassword))
}

// Login authenticates a user and returns a JWT.
func (s *AuthService) Login(ctx context.Context, email, password string) (string, *User, error) {
	user, err := s.userRepo.GetUserByEmail(ctx, email)
	if err != nil {
		return "", nil, err
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(password)); err != nil {
		return "", nil, err
	}

	token, err := s.generateToken(user.ID, user.Username)
	return token, user, err
}

func (s *AuthService) generateToken(userID int, username string) (string, error) {
	claims := jwt.MapClaims{
		"user_id":  userID,
		"username": username,
		"exp":      time.Now().Add(24 * time.Hour).Unix(),
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(s.jwtSecret))
}

// SearchUsers searches for users.
func (s *AuthService) SearchUsers(ctx context.Context, query string) ([]*User, error) {
	return s.userRepo.SearchUsers(ctx, query)
}

// ChannelService handles channel-related business logic.
type ChannelService struct {
	channelRepo ChannelRepository
}

// NewChannelService creates a new ChannelService.
func NewChannelService(channelRepo ChannelRepository) *ChannelService {
	return &ChannelService{channelRepo: channelRepo}
}

// CreateChannel creates a new channel.
func (s *ChannelService) CreateChannel(ctx context.Context, name, channelType string) (*Channel, error) {
	return s.channelRepo.CreateChannel(ctx, name, channelType)
}

// GetChannels retrieves all public channels.
func (s *ChannelService) GetChannels(ctx context.Context) ([]*Channel, error) {
	return s.channelRepo.GetChannels(ctx)
}

// GetOrCreateDM gets or creates a direct message channel.
func (s *ChannelService) GetOrCreateDM(ctx context.Context, user1ID, user2ID int) (*Channel, error) {
	return s.channelRepo.GetOrCreateDM(ctx, user1ID, user2ID)
}

// GetUserDMs retrieves all DM channels for a user.
func (s *ChannelService) GetUserDMs(ctx context.Context, userID int) ([]*Channel, error) {
	return s.channelRepo.GetUserDMs(ctx, userID)
}

// UpdateUserChannelReadStatus updates the read status for a user in a channel.
func (s *ChannelService) UpdateUserChannelReadStatus(ctx context.Context, userID, channelID, lastReadMessageID int) error {
	return s.channelRepo.UpdateUserChannelReadStatus(ctx, userID, channelID, lastReadMessageID)
}

// MessageService handles message-related business logic.
type MessageService struct {
	messageRepo MessageRepository
}

// NewMessageService creates a new MessageService.
func NewMessageService(messageRepo MessageRepository) *MessageService {
	return &MessageService{messageRepo: messageRepo}
}

// SaveMessage saves a new message.
func (s *MessageService) SaveMessage(ctx context.Context, channelID, userID int, content string) error {
	return s.messageRepo.SaveMessage(ctx, channelID, userID, content)
}

// GetMessages retrieves messages for a channel.
func (s *MessageService) GetMessages(ctx context.Context, channelID int) ([]*Message, error) {
	return s.messageRepo.GetMessages(ctx, channelID, 50) // Default limit
}

// GetLatestMessageID retrieves the latest message ID for a channel.
func (s *MessageService) GetLatestMessageID(ctx context.Context, channelID int) (int, error) {
	return s.messageRepo.GetLatestMessageID(ctx, channelID)
}
