package main

import (
	"errors"

	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5"
	"golang.org/x/crypto/bcrypt"
)

func DoTest(c *gin.Context) {
	c.JSON(200, gin.H{"result": "success"})

}

func UserLogin(c *gin.Context) {
	resp := NewAPIResponse()
	var body struct {
		UserName string `json:"userName"`
		Password string `json:"password"`
	}
	if err := c.BindJSON(&body); err != nil {
		resp.AddError(false, "Invalid JSON")
		c.JSON(400, resp)
		return
	}

	user, err := FetchUserData(c.Request.Context(), body.UserName)
	if err != nil {
		//was the user not found?
		if errors.Is(err, pgx.ErrNoRows) {
			resp.AddError(false, "Invalid login")
			c.JSON(401, resp)
			return
		} else {
			resp.AddError(false, "Unknown response from FetchUserData: "+err.Error())
			c.JSON(500, resp)
			return
		}
	}

	//validate the user's password
	err = bcrypt.CompareHashAndPassword(
		[]byte(user.PasswordHash),
		[]byte(body.Password),
	)
	if err != nil {
		resp.AddError(false, "Invalid login")
		c.JSON(401, resp)
		return
	}

	//if success
	session := SessionData{
		SessionId: GenerateSessionID(),
		UserId:    int(user.ID),
	}
	PushSessionData(c.Request.Context(), &session)
	resp.Data = gin.H{
		"sessionId": session.SessionId,
	}

	c.JSON(200, resp)
}

func UserCreate(c *gin.Context) {
	resp := NewAPIResponse()
	var body struct {
		Name     string `json:"name"`
		UserName string `json:"userName"`
		Password string `json:"password"`
	}

	if err := c.BindJSON(&body); err != nil {
		resp.AddError(false, "Invalid JSON")
		c.JSON(400, resp)
		return
	}

	result, err := DoesUsernameExistInDB(c.Request.Context(), body.UserName)
	if err != nil { //problem
		c.JSON(400, "Failure to check username")
		return
	}
	if result == true { //user exists
		resp.AddError(false, "Username not available")
		c.JSON(200, resp)
		return
	}

	//is password valid?
	if body.Password == "" {
		resp.AddError(false, "Invalid password")
		c.JSON(200, resp)
		return
	}
	hash, err := HashPassword(body.Password)

	//username is available, use it
	data := UserData{
		Name:         body.Name,
		UserName:     body.UserName,
		PasswordHash: hash,
	}
	PushUserData(c.Request.Context(), &data)

	c.JSON(200, resp)
	return
}
