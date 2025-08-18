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
	
	dbInstance := &DB{db}
	
	// Run migrations
	if err = dbInstance.RunMigrations(ctx); err != nil {
		return nil, err
	}
	
	return dbInstance, nil
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

func (db *DB) GetUserServers(ctx context.Context, userID int) ([]*models.Server, error) {
	query := `
		SELECT s.id, s.name, s.owner_id, s.created_at
		FROM servers s
		JOIN server_members sm ON s.id = sm.server_id
		WHERE sm.user_id = $1
		ORDER BY s.created_at`
	
	rows, err := db.QueryContext(ctx, query, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	
	var servers []*models.Server
	for rows.Next() {
		var server models.Server
		err := rows.Scan(&server.ID, &server.Name, &server.OwnerID, &server.CreatedAt)
		if err != nil {
			return nil, err
		}
		servers = append(servers, &server)
	}
	return servers, nil
}

func (db *DB) GetServerChannels(ctx context.Context, serverID int) ([]*models.Channel, error) {
	query := `SELECT id, server_id, name, type, created_at FROM channels WHERE server_id = $1 ORDER BY created_at`
	
	rows, err := db.QueryContext(ctx, query, serverID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	
	var channels []*models.Channel
	for rows.Next() {
		var channel models.Channel
		err := rows.Scan(&channel.ID, &channel.ServerID, &channel.Name, &channel.Type, &channel.CreatedAt)
		if err != nil {
			return nil, err
		}
		channels = append(channels, &channel)
	}
	return channels, nil
}

func (db *DB) CreateServer(ctx context.Context, name string, ownerID int) (*models.Server, error) {
	var server models.Server
	query := `INSERT INTO servers (name, owner_id) VALUES ($1, $2) RETURNING id, name, owner_id, created_at`
	err := db.QueryRowContext(ctx, query, name, ownerID).Scan(&server.ID, &server.Name, &server.OwnerID, &server.CreatedAt)
	if err != nil {
		return nil, err
	}
	
	// Auto-join owner to server
	_, err = db.ExecContext(ctx, `INSERT INTO server_members (server_id, user_id) VALUES ($1, $2)`, server.ID, ownerID)
	if err != nil {
		return nil, err
	}
	
	return &server, nil
}

func (db *DB) JoinServer(ctx context.Context, serverID, userID int) error {
	query := `INSERT INTO server_members (server_id, user_id) VALUES ($1, $2) ON CONFLICT DO NOTHING`
	_, err := db.ExecContext(ctx, query, serverID, userID)
	return err
}

func (db *DB) GetOrCreateDM(ctx context.Context, user1ID, user2ID int) (*models.Channel, error) {
	// Check if DM already exists
	query := `
		SELECT c.id, c.server_id, c.name, c.type, c.created_at
		FROM channels c
		JOIN dm_participants dp1 ON c.id = dp1.channel_id AND dp1.user_id = $1
		JOIN dm_participants dp2 ON c.id = dp2.channel_id AND dp2.user_id = $2
		WHERE c.type = 'dm'
		LIMIT 1`
	
	var channel models.Channel
	err := db.QueryRowContext(ctx, query, user1ID, user2ID).Scan(
		&channel.ID, &channel.ServerID, &channel.Name, &channel.Type, &channel.CreatedAt)
	
	if err == nil {
		return &channel, nil
	}
	
	// Create new DM channel
	err = db.QueryRowContext(ctx, 
		`INSERT INTO channels (name, type) VALUES ('dm', 'dm') RETURNING id, server_id, name, type, created_at`,
	).Scan(&channel.ID, &channel.ServerID, &channel.Name, &channel.Type, &channel.CreatedAt)

	
	if err != nil {
		return nil, err
	}
	
	// Add participants
	_, err = db.ExecContext(ctx, `INSERT INTO dm_participants (channel_id, user_id) VALUES ($1, $2), ($1, $3)`, 
		channel.ID, user1ID, user2ID)
	
	if err != nil {
		return nil, err
	}
	
	return &channel, nil
}

func (db *DB) GetUserDMs(ctx context.Context, userID int) ([]*models.Channel, error) {
	query := `
		SELECT c.id, c.server_id, c.name, c.type, c.created_at,
		       u.username as other_user
		FROM channels c
		JOIN dm_participants dp1 ON c.id = dp1.channel_id AND dp1.user_id = $1
		JOIN dm_participants dp2 ON c.id = dp2.channel_id AND dp2.user_id != $1
		JOIN users u ON dp2.user_id = u.id
		WHERE c.type = 'dm'
		ORDER BY c.created_at DESC`
	
	rows, err := db.QueryContext(ctx, query, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	
	var channels []*models.Channel
	for rows.Next() {
		var channel models.Channel
		var otherUser string
		err := rows.Scan(&channel.ID, &channel.ServerID, &channel.Name, &channel.Type, &channel.CreatedAt, &otherUser)
		if err != nil {
			return nil, err
		}
		channel.Name = otherUser // Use other user's name as DM name
		channels = append(channels, &channel)
	}
	return channels, nil
}