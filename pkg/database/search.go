package database

import (
	"context"
	"github.com/divyeshmangla/nexus/internal/models"
)

func (db *DB) SearchUsers(ctx context.Context, query string) ([]*models.User, error) {
	searchQuery := "%" + query + "%"
	sql := `SELECT id, username, email, created_at FROM users WHERE username ILIKE $1 LIMIT 10`
	
	rows, err := db.QueryContext(ctx, sql, searchQuery)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	
	var users []*models.User
	for rows.Next() {
		var user models.User
		err := rows.Scan(&user.ID, &user.Username, &user.Email, &user.CreatedAt)
		if err != nil {
			return nil, err
		}
		users = append(users, &user)
	}
	return users, nil
}