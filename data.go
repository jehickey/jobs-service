package main

import (
	"crypto/rand"
	"encoding/hex"
	"log"
	"time"

	"github.com/gin-gonic/gin"
	"golang.org/x/crypto/bcrypt"
)

type SessionData struct {
	SessionId string    `json:"sessionId"`
	UserId    int       `json:"userId"`
	Created   time.Time `json:"created"`
	Expires   time.Time `json:"expires"`
	Name      string    `json:"name"`
	UserName  string    `json:"userName"`
}

type UserData struct {
	ID           int64     `json:"id"`
	Name         string    `json:"name"`
	UserName     string    `json:"userName"`
	PasswordHash string    `json:"passwordHash"`
	Created      time.Time `json:"created"`
	Updated      time.Time `json:"updated"`
	LastLogin    time.Time `json:"lastLogin"`
}

type ApplicationData struct {
	ID           int       `json:"id"`
	UserId       int       `json:"userId"`
	Position     string    `json:"position"`
	StatusId     int       `json:"statusId"`
	Organization int       `json:"organization"`
	DateApplied  time.Time `json:"dateApplied"`
	LastResponse time.Time `json:"lateResponse"`
	Url          string    `json:"url"`
	SiteUser     string    `json:"siteUser"`
	SitePass     string    `json:"sitePass"`
	SourceId     int       `json:"sourceId"`
	JobPosting   string    `json:"jobPosting"`
	Notes        string    `json:"notes"`
	Created      time.Time `json:"created"`
	Updated      time.Time `json:"updated"`
}

func GenerateSessionID() string {
	b := make([]byte, 16)
	_, err := rand.Read(b)
	if err != nil {
		return ""
	}
	return hex.EncodeToString(b) // 64-char hex string
}

func HashPassword(password string) (string, error) {
	bytes, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	return string(bytes), err
}

func GetUserFromSession(c *gin.Context) int {
	sessionId := c.GetHeader("X-SessionId")
	session, err := FetchSessionData(c.Request.Context(), sessionId)
	if err != nil {
		log.Printf("Session check failed: %v, %v", sessionId, err)
		return 0
	}
	return session.UserId
}
