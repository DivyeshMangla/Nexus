package core

import (
	"context"
	"time"

	"golang.org/x/crypto/bcrypt"
	"github.com/golang-jwt/jwt/v5"
)

// Service handles business logic
type Service struct {
	repo      Repository
	jwtSecret string
}

// NewService creates a new service
func NewService(repo Repository, jwtSecret string) *Service {
	return &Service{
		repo:      repo,
		jwtSecret: jwtSecret,
	}
}

// Auth methods
func (s *Service) Register(ctx context.Context, req SignupRequest) error {
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.Password), 12)
	if err != nil {
		return err
	}
	return s.repo.CreateUser(ctx, req.Username, req.Email, string(hashedPassword))
}

func (s *Service) Login(ctx context.Context, req LoginRequest) (string, *User, error) {
	user, err := s.repo.GetUserByEmail(ctx, req.Email)
	if err != nil {
		return "", nil, err
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(req.Password)); err != nil {
		return "", nil, err
	}

	token, err := s.generateToken(user.ID, user.Username)
	return token, user, err
}

func (s *Service) generateToken(userID int, username string) (string, error) {
	claims := jwt.MapClaims{
		"user_id":  userID,
		"username": username,
		"exp":      time.Now().Add(24 * time.Hour).Unix(),
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(s.jwtSecret))
}

// Channel methods
func (s *Service) CreateChannel(ctx context.Context, req CreateChannelRequest) (*Channel, error) {
	return s.repo.CreateChannel(ctx, req.Name, req.Type)
}

func (s *Service) GetChannels(ctx context.Context) ([]*Channel, error) {
	return s.repo.GetChannels(ctx)
}

func (s *Service) GetOrCreateDM(ctx context.Context, user1ID, user2ID int) (*Channel, error) {
	return s.repo.GetOrCreateDM(ctx, user1ID, user2ID)
}

// Message methods
func (s *Service) SaveMessage(ctx context.Context, channelID, userID int, content string) error {
	return s.repo.SaveMessage(ctx, channelID, userID, content)
}

func (s *Service) GetMessages(ctx context.Context, channelID int) ([]*Message, error) {
	return s.repo.GetMessages(ctx, channelID, 50)
}

func (s *Service) SearchUsers(ctx context.Context, query string) ([]*User, error) {
	return s.repo.SearchUsers(ctx, query)
}