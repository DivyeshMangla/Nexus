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
		`CREATE TABLE IF NOT EXISTS user_channel_read_status (
			user_id INTEGER NOT NULL,
			channel_id INTEGER NOT NULL,
			last_read_message_id INTEGER NOT NULL,
			PRIMARY KEY (user_id, channel_id)
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
	fmt.Printf("GetOrCreateDM called with user1ID=%d, user2ID=%d\n", user1ID, user2ID)
	
	// Check if DM exists
	var channelID int
	err := db.QueryRowContext(ctx, `
		SELECT dp.channel_id
		FROM dm_participants dp
		JOIN channels c ON dp.channel_id = c.id
		WHERE dp.user_id IN ($1, $2) AND c.type = 'dm'
		GROUP BY dp.channel_id
		HAVING COUNT(DISTINCT dp.user_id) = 2
	`, user1ID, user2ID).Scan(&channelID)

	if err == nil {
		fmt.Printf("Found existing DM with channelID=%d\n", channelID)
		var channel core.Channel
		err = db.QueryRowContext(ctx,
			`SELECT id, name, type FROM channels WHERE id = $1`,
			channelID).Scan(&channel.ID, &channel.Name, &channel.Type)
		if err != nil {
			fmt.Printf("Error getting channel details: %v\n", err)
		}
		return &channel, err
	}

	// Only create new DM if no rows found, otherwise return the error
	if err != sql.ErrNoRows {
		fmt.Printf("Unexpected error checking for existing DM: %v\n", err)
		return nil, err
	}

	fmt.Printf("No existing DM found, creating new one\n")
	
	// Create new DM
	tx, err := db.BeginTx(ctx, nil)
	if err != nil {
		fmt.Printf("Error starting transaction: %v\n", err)
		return nil, err
	}
	defer tx.Rollback()

	var channel core.Channel
	err = tx.QueryRowContext(ctx,
		`INSERT INTO channels (name, type) VALUES ('dm', 'dm') RETURNING id, name, type`,
	).Scan(&channel.ID, &channel.Name, &channel.Type)
	if err != nil {
		fmt.Printf("Error creating channel: %v\n", err)
		return nil, err
	}

	fmt.Printf("Created channel with ID=%d\n", channel.ID)

	_, err = tx.ExecContext(ctx,
		`INSERT INTO dm_participants (channel_id, user_id) VALUES ($1, $2), ($1, $3)`,
		channel.ID, user1ID, user2ID)
	if err != nil {
		fmt.Printf("Error inserting participants: %v\n", err)
		return nil, err
	}

	fmt.Printf("Inserted participants successfully\n")

	err = tx.Commit()
	if err != nil {
		fmt.Printf("Error committing transaction: %v\n", err)
		return nil, err
	}

	fmt.Printf("Transaction committed successfully\n")
	return &channel, nil
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

func (db *DB) GetUserDMs(ctx context.Context, userID int) ([]*core.Channel, error) {
	rows, err := db.QueryContext(ctx, `
		SELECT DISTINCT c.id, c.name, c.type, u.username
		FROM channels c
		JOIN dm_participants dp1 ON c.id = dp1.channel_id
		JOIN dm_participants dp2 ON c.id = dp2.channel_id
		JOIN users u ON dp2.user_id = u.id
		WHERE c.type = 'dm' AND dp1.user_id = $1 AND dp2.user_id != $1
		ORDER BY c.id DESC
	`, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var dms []*core.Channel
	for rows.Next() {
		var dm core.Channel
		var username string
		if err := rows.Scan(&dm.ID, &dm.Name, &dm.Type, &username); err != nil {
			return nil, err
		}
		// Use the other user's name as DM name
		dm.Name = username
		dms = append(dms, &dm)
	}
	return dms, nil
}

func (db *DB) UpdateUserChannelReadStatus(ctx context.Context, userID, channelID, lastReadMessageID int) error {
	_, err := db.ExecContext(ctx, `
		INSERT INTO user_channel_read_status (user_id, channel_id, last_read_message_id)
		VALUES ($1, $2, $3)
		ON CONFLICT (user_id, channel_id) DO UPDATE SET last_read_message_id = $3
	`, userID, channelID, lastReadMessageID)
	return err
}

func (db *DB) GetLatestMessageID(ctx context.Context, channelID int) (int, error) {
	var messageID int
	err := db.QueryRowContext(ctx, `
		SELECT id FROM messages
		WHERE channel_id = $1
		ORDER BY created_at DESC
		LIMIT 1
	`, channelID).Scan(&messageID)
	if err == sql.ErrNoRows {
		return 0, nil // No messages in channel, return 0 or handle as appropriate
	}
	return messageID, err
}

func (db *DB) Close() error {
	return db.DB.Close()
}