package main

import (
	"errors"
	"log"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5"
	"golang.org/x/crypto/bcrypt"
)

func DoTest(c *gin.Context) {
	c.JSON(200, gin.H{"result": "success"})

}

func GetSessionInfo(c *gin.Context) {
	sessionId := c.GetHeader("X-SessionId")
	if sessionId == "" {
		c.JSON(401, gin.H{"error": "No active session"})
		return
	}

	session, err := FetchSessionData(c.Request.Context(), sessionId)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			c.JSON(401, gin.H{"error": "Session invalid"})
			return
		}
		c.JSON(500, gin.H{"error": "Unknown error: " + err.Error()})
	}

	c.JSON(200, gin.H{
		"userId":   session.UserId,
		"name":     session.Name,
		"userName": session.UserName,
	})
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

	log.Printf("Requesting login for %v pw %v", body.UserName, body.Password)

	user, err := FetchUserData(c.Request.Context(), body.UserName)
	if err != nil {
		//was the user not found?
		if errors.Is(err, pgx.ErrNoRows) {
			//resp.AddError(false, "Invalid login")
			c.JSON(401, "rejected")
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
	if err != nil { //password does not match
		c.JSON(401, "rejected")
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

	c.JSON(200, session.SessionId)
}

func UserLogout(c *gin.Context) {
	//do stuff to kill their session id
	sessionId := c.GetHeader("X-SessionId")
	if sessionId == "" {
		c.JSON(401, gin.H{"error": "No active session"})
		return
	}

	_, err := FetchSessionData(c.Request.Context(), sessionId)
	if err == nil {
		//db command to terminate the session
	}
	c.Status(200)
}

// return true if user is found
func CheckUserExists(c *gin.Context) {
	username := c.DefaultQuery("username", "")

	result, err := DoesUsernameExistInDB(c.Request.Context(), username)
	if err != nil { //problem
		c.JSON(500, err)
		return
	}
	log.Printf("username check: result  = %+v", result)
	c.JSON(200, result)
	return
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

	log.Printf("Requesting user creation for %v (%v) pw %v", body.UserName, body.Name, body.Password)
	//validate username

	result, err := DoesUsernameExistInDB(c.Request.Context(), body.UserName)
	if err != nil { //problem
		c.JSON(400, "Failure to check username")
		return
	}
	if result == true { //user exists
		resp.AddError(false, "Username not available")
		log.Printf("Name already exists")
		c.JSON(400, resp)
		return
	}

	//is password valid?
	if body.Password == "" {
		resp.AddError(false, "Invalid password")
		c.JSON(400, resp)
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

/// APPLICATIONS

func CreateApplication(c *gin.Context) {
	log.Printf("creating application")
	userid := GetUserFromSession(c)
	if userid == 0 {
		c.Status(401)
		return
	}
	appId, err := CreateApplicationInDB(c.Request.Context(), int64(userid))
	if err != nil {
		c.Status(500)
		log.Printf("Got error while creating application: %v", err)
		return
	}
	log.Printf("Got id while creating application: %v", appId)
	c.JSON(200, gin.H{"applicationId": appId})
}

func GetApplication(c *gin.Context) {
	userid := GetUserFromSession(c)
	if userid == 0 {
		c.Status(401)
		return
	}
	appId, err := strconv.Atoi(c.Param("id"))
	if appId == 0 || err != nil {
		c.Status(401)
		return
	}
	//get the app
	data, err := FetchApplication(c.Request.Context(), appId, userid)
	c.JSON(200, data) //return the app
}

func UpdateApplication(c *gin.Context) {
	userId := GetUserFromSession(c)
	log.Printf("Update App Request: user:%v session:", userId)
	if userId == 0 {
		c.Status(401)
		return
	}
	appId, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		appId = 0
	}
	fieldName := c.Param("field")

	var body struct {
		Value string `json:"value"`
	}
	if err := c.BindJSON(&body); err != nil {
		c.Status(400)
		return
	}

	log.Printf("Update App Request: app:%v user:%v %v=%v", appId, userId, fieldName, body.Value)

	//does the user own this app?
	own := VerifyApplicationOwnership(c.Request.Context(), appId, userId)
	if own == false {
		log.Printf("Update App ownership fail: app:%v is not owned by user:%v", appId, userId)
		c.Status(401)
		return
	}

	log.Printf("Updating Application: app:%v user:%v %v=%v", appId, userId, fieldName, body.Value)
	//update a field
	err = UpdateApplicationField(c.Request.Context(), appId, fieldName, body.Value)
	if err != nil { //problem with editing
		log.Printf("Update failed: %v", err)
		c.Status(500)
		return
	}

	c.Status(200)
}

func DeleteApplication(c *gin.Context) {
	userid := GetUserFromSession(c)
	if userid == 0 {
		c.Status(401)
		return
	}
	//does the user own this app?
	//delete it
}

func GetApplicationList(c *gin.Context) {
	userId := GetUserFromSession(c)
	if userId == 0 {
		c.Status(401)
		return
	}
	list, err := FetchApplicationList(c.Request.Context(), userId)
	if err != nil {
		log.Printf("FetchApplicationList failed for user %d: %v", userId, err)
		c.JSON(500, gin.H{"error": err.Error()})
		return
	}
	c.JSON(200, list)
}
