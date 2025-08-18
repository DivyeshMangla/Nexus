package database

import (
	"context"
	"database/sql"
	"fmt"
	"time"
	_ "github.com/lib/pq"
	"github.com/divyeshmangla/nexus/internal/models"
)

type DB struct {
	*sql.DB
}

func Connect(host, port, user, password, dbname string) (Repository, error) {
	psqlInfo := fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=disable",
		host, port, user, password, dbname)
	
	db, err := sql.Open("postgres", psqlInfo)
	if err != nil {
		return nil, err
	}
	
	// Connection pool settings
	db.SetMaxOpenConns(25)
	db.SetMaxIdleConns(5)
	db.SetConnMaxLifetime(5 * time.Minute)
	
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	
	if err = db.PingContext(ctx); err != nil {
		return nil, err
	}
	
	return &DB{db}, nil
}

func (db *DB) CreateUser(ctx context.Context, username, email, hashedPassword string) error {
	query := `INSERT INTO users (username, email, password) VALUES ($1, $2, $3)`
	_, err := db.ExecContext(ctx, query, username, email, hashedPassword)
	return err
}

func (db *DB) GetUserByEmail(ctx context.Context, email string) (*models.User, error) {
	query := `SELECT id, username, email, password, created_at FROM users WHERE email = $1`
	var user models.User
	err := db.QueryRowContext(ctx, query, email).Scan(
		&user.ID, &user.Username, &user.Email, &user.Password, &user.CreatedAt)
	if err != nil {
		return nil, err
	}
	return &user, nil
}

func (db *DB) SaveMessage(ctx context.Context, channelID, userID int, content, username string) error {
	query := `INSERT INTO messages (channel_id, user_id, content) VALUES ($1, $2, $3)`
	_, err := db.ExecContext(ctx, query, channelID, userID, content)
	return err
}

func (db *DB) GetRecentMessages(ctx context.Context, channelID int, limit int) ([]*models.Message, error) {
	query := `
		SELECT m.id, m.channel_id, m.user_id, m.content, m.created_at, u.username
		FROM messages m
		JOIN users u ON m.user_id = u.id
		WHERE m.channel_id = $1
		ORDER BY m.created_at DESC
		LIMIT $2`
	
	rows, err := db.QueryContext(ctx, query, channelID, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	
	var messages []*models.Message
	for rows.Next() {
		var msg models.Message
		err := rows.Scan(&msg.ID, &msg.ChannelID, &msg.UserID, &msg.Content, &msg.CreatedAt, &msg.Username)
		if err != nil {
			return nil, err
		}
		messages = append(messages, &msg)
	}
	
	// Reverse to get chronological order
	for i, j := 0, len(messages)-1; i < j; i, j = i+1, j-1 {
		messages[i], messages[j] = messages[j], messages[i]
	}
	
	return messages, nil
}