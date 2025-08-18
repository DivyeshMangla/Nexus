package database

import (
	"context"
	"log"
)

func (db *DB) RunMigrations(ctx context.Context) error {
	migrations := []string{
		// Create users table
		`CREATE TABLE IF NOT EXISTS users (
			id SERIAL PRIMARY KEY,
			username VARCHAR(50) UNIQUE NOT NULL,
			email VARCHAR(100) UNIQUE NOT NULL,
			password VARCHAR(255) NOT NULL,
			created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
		)`,
		
		// Drop and recreate servers table with correct columns
		`DROP TABLE IF EXISTS servers CASCADE`,
		`CREATE TABLE servers (
			id SERIAL PRIMARY KEY,
			name VARCHAR(100) NOT NULL,
			owner_id INTEGER NOT NULL,
			created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
		)`,
		
		// Drop and recreate channels table
		`DROP TABLE IF EXISTS channels CASCADE`,
		`CREATE TABLE channels (
			id SERIAL PRIMARY KEY,
			server_id INTEGER,
			name VARCHAR(100) NOT NULL,
			type VARCHAR(20) DEFAULT 'text',
			created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
		)`,
		
		// Create server_members table
		`CREATE TABLE IF NOT EXISTS server_members (
			id SERIAL PRIMARY KEY,
			server_id INTEGER NOT NULL,
			user_id INTEGER NOT NULL,
			joined_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
			UNIQUE(server_id, user_id)
		)`,
		
		// Create dm_participants table
		`CREATE TABLE IF NOT EXISTS dm_participants (
			id SERIAL PRIMARY KEY,
			channel_id INTEGER NOT NULL,
			user_id INTEGER NOT NULL,
			UNIQUE(channel_id, user_id)
		)`,
		
		// Create read status table
		`CREATE TABLE IF NOT EXISTS channel_read_status (
			id SERIAL PRIMARY KEY,
			user_id INTEGER NOT NULL,
			channel_id INTEGER NOT NULL,
			last_read_at TIMESTAMP NOT NULL,
			UNIQUE(user_id, channel_id)
		)`,
		
		// Recreate messages table with proper foreign key
		`DROP TABLE IF EXISTS messages CASCADE`,
		`CREATE TABLE messages (
			id SERIAL PRIMARY KEY,
			channel_id INTEGER NOT NULL DEFAULT 1,
			user_id INTEGER NOT NULL,
			content TEXT NOT NULL,
			created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
		)`,
		
		// Insert default user
		`INSERT INTO users (id, username, email, password) VALUES (1, 'system', 'system@nexus.com', 'dummy') ON CONFLICT (id) DO NOTHING`,
		
		// Insert default server and channel
		`INSERT INTO servers (name, owner_id) VALUES ('Nexus Server', 1)`,
		`INSERT INTO channels (server_id, name, type) VALUES (1, 'general', 'text')`,
	}
	
	for _, migration := range migrations {
		_, err := db.ExecContext(ctx, migration)
		if err != nil {
			log.Printf("Migration failed: %v", err)
			return err
		}
	}
	
	log.Println("Migrations completed successfully")
	return nil
}