package database

import (
	"context"
	"time"
)

func (db *DB) MarkChannelRead(ctx context.Context, userID, channelID int) error {
	query := `
		INSERT INTO channel_read_status (user_id, channel_id, last_read_at) 
		VALUES ($1, $2, $3)
		ON CONFLICT (user_id, channel_id) 
		DO UPDATE SET last_read_at = $3`
	
	_, err := db.ExecContext(ctx, query, userID, channelID, time.Now())
	return err
}

func (db *DB) GetUnreadChannels(ctx context.Context, userID int) ([]int, error) {
	query := `
		SELECT DISTINCT m.channel_id
		FROM messages m
		LEFT JOIN channel_read_status crs ON crs.user_id = $1 AND crs.channel_id = m.channel_id
		LEFT JOIN dm_participants dp ON dp.channel_id = m.channel_id AND dp.user_id = $1
		WHERE (m.channel_id = 1 OR dp.user_id IS NOT NULL)
		AND (crs.last_read_at IS NULL OR m.created_at > crs.last_read_at)
		AND m.user_id != $1`
	
	rows, err := db.QueryContext(ctx, query, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	
	var unreadChannels []int
	for rows.Next() {
		var channelID int
		if err := rows.Scan(&channelID); err != nil {
			return nil, err
		}
		unreadChannels = append(unreadChannels, channelID)
	}
	
	return unreadChannels, nil
}