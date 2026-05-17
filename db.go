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
		//"SELECT sessionid, userid, expires FROM sessions WHERE sessionid=$1", sessionId,
		"SELECT sessionid, userid, expires, a.name, a.username FROM sessions inner join users a on a.id = userid WHERE sessionid=$1 AND expires > Now() AND active=true", sessionId,
	).Scan(&data.SessionId, &data.UserId, &data.Expires, &data.Name, &data.UserName)
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

/// APPLICATIONS

func FetchApplication(ctx context.Context, id int, userId int) (ApplicationData, error) {
	data := ApplicationData{}

	sql := `SELECT id, userid, COALESCE(organizationid,0), position, 
			dateapplied,
			lastresponse,
			statusid, jobposting, notes, url, siteuser,
			sourceid, created, updated FROM applications
			WHERE id=$1 AND userid=$2`

	err := DB.QueryRow(ctx, sql, id, userId).Scan(
		&data.ID, &data.UserId, &data.Organization, &data.Position,
		&data.DateApplied, &data.LastResponse, &data.StatusId,
		&data.JobPosting, &data.Notes, &data.Url, &data.SiteUser,
		&data.SourceId, &data.Created, &data.Updated)
	if err != nil {
		log.Printf("DB Error on FetchApplication (user=%v : id=%v): %v", data.UserId, data.ID, err)
	}
	return data, err
}

func CreateApplicationInDB(ctx context.Context, userId int64) (int64, error) {
	//insert into applications (userid) values (6) returning id;
	var appId int64 = 0
	sql := `INSERT INTO applications (userid) values ($1) returning id`
	err := DB.QueryRow(ctx, sql, userId).Scan(&appId)
	if err != nil {
		log.Printf("DB Error while creating application: %v", err.Error())
	}
	return appId, err
}

// Returns TRUE if this application is owned by this user
func VerifyApplicationOwnership(ctx context.Context, appId int, userId int) bool {
	var count int
	err := DB.QueryRow(ctx, "SELECT COUNT(*) FROM applications WHERE id=$1 AND userid=$2", appId, userId).Scan(&count)
	if err != nil {
		log.Printf("Error verifying application ownership: %+v", err)
		return false
	}
	return count > 0
}

var editableApplicationFields = map[string]string{
	"position":     "position",
	"organization": "organization",
	"statusId":     "statusid",
	"sourceId":     "sourceid",
	"jobPosting":   "jobposting",
	"notes":        "notes",
	"url":          "url",
	"siteUser":     "siteuser",
	"sitePass":     "sitepass",
	"dateApplied":  "dateapplied",
	"lastResponse": "lastresponse",
}

func UpdateApplicationField(ctx context.Context, appId int, field string, value string) error {
	// Validate field
	column, ok := editableApplicationFields[field]
	if !ok {
		return fmt.Errorf("field %s is not editable", field)
	}

	// Build SQL safely
	query := fmt.Sprintf(`
        UPDATE applications
        SET %s = $1, updated = NOW()
        WHERE id = $2
    `, column)

	// Execute
	_, err := DB.Exec(ctx, query, value, appId)
	return err
}

func FetchApplicationList(ctx context.Context, userId int) ([]ApplicationData, error) {
	apps := []ApplicationData{}

	sql := `SELECT id, userid, COALESCE(position, ''), statusid
		    FROM applications
			WHERE userid=$1
			ORDER BY created ASC`

	rows, err := DB.Query(ctx, sql, userId)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	for rows.Next() {
		var a ApplicationData
		err := rows.Scan(&a.ID, &a.UserId, &a.Position, &a.StatusId)
		if err != nil {
			return nil, err
		}
		apps = append(apps, a)
	}
	return apps, rows.Err()
}
