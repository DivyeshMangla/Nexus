package database

import (
	"database/sql"
	"fmt"
	_ "github.com/lib/pq"
)

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
	_, err := db.Exec("INSERT INTO users (username, email, password) VALUES ($1, $2, $3)",
		username, email, hashedPassword)
	return err
}

func (db *DB) GetUserByEmail(email string) (int, string, string, error) {
	var id int
	var username, password string
	err := db.QueryRow("SELECT id, username, password FROM users WHERE email = $1", email).
		Scan(&id, &username, &password)
	return id, username, password, err
}