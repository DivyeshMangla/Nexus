package database

import (
	"context"
	"github.com/divyeshmangla/nexus/internal/models"
)

type Repository interface {
	CreateUser(ctx context.Context, username, email, hashedPassword string) error
	GetUserByEmail(ctx context.Context, email string) (*models.User, error)
	Close() error
}