package main

import (
	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
)

func main() {
	if err := InitDB(); err != nil {
		panic(err)
	}

	r := gin.Default()
	r.SetTrustedProxies(nil)
	r.Use(cors.New(cors.Config{
		AllowOrigins:     []string{"http://jobs.ehickey.com:4200"},
		AllowMethods:     []string{"GET", "POST", "PUT", "DELETE", "OPTIONS", "PATCH"},
		AllowHeaders:     []string{"Content-Type", "Authorization", "X-SessionId"},
		AllowCredentials: true,
	}))

	r.GET("/test", DoTest)

	r.POST("/login", UserLogin)
	r.POST("/logout", UserLogout)
	r.POST("/users", UserCreate)
	r.GET("/users/exists", CheckUserExists)
	r.GET("/session", GetSessionInfo)

	r.GET("/applications", GetApplicationList)
	r.POST("/applications", CreateApplication)
	r.GET("/applications/:id", GetApplication)
	r.PATCH("/applications/:id/:field", UpdateApplication)
	r.DELETE("/applications/:id", DeleteApplication)

	r.Run("0.0.0.0:8081")
}
