package main

import (
	"crypto/rand"
	"encoding/hex"
	"time"

	"golang.org/x/crypto/bcrypt"
)

type SessionData struct {
	SessionId string    `json:"sessionId"`
	UserId    int       `json:"userId"`
	Created   time.Time `json:"created"`
	Expires   time.Time `json:"expires"`
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

func GenerateSessionID() string {
	b := make([]byte, 32) // 256 bits
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
