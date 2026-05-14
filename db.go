package main

import (
	"context"
	"fmt"
	"log"

	"github.com/jackc/pgx/v5/pgxpool"
)

var DB *pgxpool.Pool

func InitDB() error {
	url := "postgres://postgres:kookookachoo@localhost:5432/jobsite"
	pool, err := pgxpool.New(context.Background(), url)
	if err != nil {
		return fmt.Errorf("Unable to create connection pool: %w", err)
	}

	//test connection
	err = pool.Ping(context.Background())
	if err != nil {
		return fmt.Errorf("Unable to connect to database: %w", err)
	}

	DB = pool
	return nil
}

func FetchUserData(ctx context.Context, username string) (UserData, error) {
	data := UserData{}

	err := DB.QueryRow(ctx,
		"SELECT id, username, passwordhash FROM users WHERE username=$1", username,
	).Scan(&data.ID, &data.UserName, &data.PasswordHash)
	return data, err
}

func FetchSessionData(ctx context.Context, sessionId string) (SessionData, error) {
	data := SessionData{}
	err := DB.QueryRow(ctx,
		"SELECT sessionid, userid, expires FROM sessions WHERE sessionid=$1", sessionId,
	).Scan(&data.SessionId, &data.UserId, &data.Expires)
	return data, err
}

func PushSessionData(ctx context.Context, data *SessionData) error {
	sql := `
	INSERT INTO sessions
	(sessionid, userid, expires)
	VALUES ($1, $2, NOW()+INTERVAL '24 hours')
	`
	_, err := DB.Exec(ctx, sql, data.SessionId, data.UserId)
	return err
}

func DoesUsernameExistInDB(ctx context.Context, username string) (bool, error) {
	var count int
	err := DB.QueryRow(ctx, "SELECT COUNT(*) FROM users WHERE username=$1", username).Scan(&count)
	if err != nil {
		log.Printf("Error getting username count: %+v", err)
		return false, err
	}
	return count > 0, nil
}

func PushUserData(ctx context.Context, data *UserData) error {
	sql := `
	INSERT INTO users
	(name, username, passwordHash)
	VALUES ($1, $2, $3)
	`
	_, err := DB.Exec(ctx, sql, data.Name, data.UserName, data.PasswordHash)
	return err
}
