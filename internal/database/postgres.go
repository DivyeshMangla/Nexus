package database

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	_ "github.com/lib/pq"
	"github.com/divyeshmangla/nexus/internal/core"
)

type DB struct {
	*sql.DB
}

func Connect(host, port, user, password, dbname string) (*DB, error) {
	dsn := fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=disable",
		host, port, user, password, dbname)

	db, err := sql.Open("postgres", dsn)
	if err != nil {
		return nil, err
	}

	db.SetMaxOpenConns(25)
	db.SetMaxIdleConns(5)
	db.SetConnMaxLifetime(5 * time.Minute)

	if err = db.Ping(); err != nil {
		return nil, err
	}

	instance := &DB{db}
	if err = instance.migrate(); err != nil {
		return nil, err
	}

	return instance, nil
}

func (db *DB) migrate() error {
	queries := []string{
		`CREATE TABLE IF NOT EXISTS users (
			id SERIAL PRIMARY KEY,
			username VARCHAR(50) UNIQUE NOT NULL,
			email VARCHAR(100) UNIQUE NOT NULL,
			password VARCHAR(255) NOT NULL,
			created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
		)`,
		`CREATE TABLE IF NOT EXISTS channels (
			id SERIAL PRIMARY KEY,
			name VARCHAR(100) NOT NULL,
			type VARCHAR(20) DEFAULT 'general',
			created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
		)`,
		`CREATE TABLE IF NOT EXISTS messages (
			id SERIAL PRIMARY KEY,
			channel_id INTEGER NOT NULL DEFAULT 1,
			user_id INTEGER NOT NULL,
			content TEXT NOT NULL,
			created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
		)`,
		`CREATE TABLE IF NOT EXISTS dm_participants (
			channel_id INTEGER NOT NULL,
			user_id INTEGER NOT NULL,
			PRIMARY KEY (channel_id, user_id)
		)`,
		`INSERT INTO channels (id, name, type) VALUES (1, 'general', 'general') ON CONFLICT DO NOTHING`,
	}

	for _, query := range queries {
		if _, err := db.Exec(query); err != nil {
			return err
		}
	}
	return nil
}

// User operations
func (db *DB) CreateUser(ctx context.Context, username, email, hashedPassword string) error {
	_, err := db.ExecContext(ctx,
		`INSERT INTO users (username, email, password) VALUES ($1, $2, $3)`,
		username, email, hashedPassword)
	return err
}

func (db *DB) GetUserByEmail(ctx context.Context, email string) (*core.User, error) {
	var user core.User
	err := db.QueryRowContext(ctx,
		`SELECT id, username, email, password FROM users WHERE email = $1`,
		email).Scan(&user.ID, &user.Username, &user.Email, &user.Password)
	return &user, err
}

func (db *DB) SearchUsers(ctx context.Context, query string) ([]*core.User, error) {
	rows, err := db.QueryContext(ctx,
		`SELECT id, username, email FROM users WHERE username ILIKE $1 LIMIT 10`,
		"%"+query+"%")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var users []*core.User
	for rows.Next() {
		var user core.User
		if err := rows.Scan(&user.ID, &user.Username, &user.Email); err != nil {
			return nil, err
		}
		users = append(users, &user)
	}
	return users, nil
}

// Channel operations
func (db *DB) CreateChannel(ctx context.Context, name, channelType string) (*core.Channel, error) {
	var channel core.Channel
	err := db.QueryRowContext(ctx,
		`INSERT INTO channels (name, type) VALUES ($1, $2) RETURNING id, name, type`,
		name, channelType).Scan(&channel.ID, &channel.Name, &channel.Type)
	return &channel, err
}

func (db *DB) GetChannels(ctx context.Context) ([]*core.Channel, error) {
	rows, err := db.QueryContext(ctx, `SELECT id, name, type FROM channels WHERE type = 'general'`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var channels []*core.Channel
	for rows.Next() {
		var channel core.Channel
		if err := rows.Scan(&channel.ID, &channel.Name, &channel.Type); err != nil {
			return nil, err
		}
		channels = append(channels, &channel)
	}
	return channels, nil
}

func (db *DB) GetOrCreateDM(ctx context.Context, user1ID, user2ID int) (*core.Channel, error) {
	// Check if DM exists
	var channelID int
	err := db.QueryRowContext(ctx, `
		SELECT dp1.channel_id FROM dm_participants dp1
		JOIN dm_participants dp2 ON dp1.channel_id = dp2.channel_id
		WHERE dp1.user_id = $1 AND dp2.user_id = $2
	`, user1ID, user2ID).Scan(&channelID)

	if err == nil {
		var channel core.Channel
		err = db.QueryRowContext(ctx,
			`SELECT id, name, type FROM channels WHERE id = $1`,
			channelID).Scan(&channel.ID, &channel.Name, &channel.Type)
		return &channel, err
	}

	// Create new DM
	tx, err := db.BeginTx(ctx, nil)
	if err != nil {
		return nil, err
	}
	defer tx.Rollback()

	var channel core.Channel
	err = tx.QueryRowContext(ctx,
		`INSERT INTO channels (name, type) VALUES ('dm', 'dm') RETURNING id, name, type`,
	).Scan(&channel.ID, &channel.Name, &channel.Type)
	if err != nil {
		return nil, err
	}

	_, err = tx.ExecContext(ctx,
		`INSERT INTO dm_participants (channel_id, user_id) VALUES ($1, $2), ($1, $3)`,
		channel.ID, user1ID, user2ID)
	if err != nil {
		return nil, err
	}

	return &channel, tx.Commit()
}

// Message operations
func (db *DB) SaveMessage(ctx context.Context, channelID, userID int, content string) error {
	_, err := db.ExecContext(ctx,
		`INSERT INTO messages (channel_id, user_id, content) VALUES ($1, $2, $3)`,
		channelID, userID, content)
	return err
}

func (db *DB) GetMessages(ctx context.Context, channelID int, limit int) ([]*core.Message, error) {
	rows, err := db.QueryContext(ctx, `
		SELECT m.id, m.channel_id, m.user_id, m.content, m.created_at, u.username
		FROM messages m
		JOIN users u ON m.user_id = u.id
		WHERE m.channel_id = $1
		ORDER BY m.created_at DESC
		LIMIT $2
	`, channelID, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var messages []*core.Message
	for rows.Next() {
		var msg core.Message
		if err := rows.Scan(&msg.ID, &msg.ChannelID, &msg.UserID, &msg.Content, &msg.CreatedAt, &msg.Username); err != nil {
			return nil, err
		}
		messages = append(messages, &msg)
	}

	// Reverse for chronological order
	for i, j := 0, len(messages)-1; i < j; i, j = i+1, j-1 {
		messages[i], messages[j] = messages[j], messages[i]
	}

	return messages, nil
}