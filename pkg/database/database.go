package database

import (
	"database/sql"
	"fmt"
	_ "github.com/lib/pq"
	"github.com/divyeshmangla/nexus/internal/models"
)

type User = models.User

type DB struct {
	*sql.DB
}

func Connect(host, port, user, password, dbname string) (*DB, error) {
	psqlInfo := fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=disable",
		host, port, user, password, dbname)
	
	db, err := sql.Open("postgres", psqlInfo)
	if err != nil {
		return nil, err
	}
	
	if err = db.Ping(); err != nil {
		return nil, err
	}
	
	return &DB{db}, nil
}

func (db *DB) CreateUser(username, email, hashedPassword string) error {
	query := `INSERT INTO users (username, email, password) VALUES ($1, $2, $3)`
	_, err := db.Exec(query, username, email, hashedPassword)
	return err
}

func (db *DB) GetUserByEmail(email string) (*User, error) {
	query := `SELECT id, username, email, password, created_at FROM users WHERE email = $1`
	var user User
	err := db.QueryRow(query, email).Scan(
		&user.ID, &user.Username, &user.Email, &user.Password, &user.CreatedAt)
	if err != nil {
		return nil, err
	}
	return &user, nil
}