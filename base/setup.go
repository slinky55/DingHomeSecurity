package main

import (
	"github.com/gin-contrib/sessions"
	"github.com/gin-contrib/sessions/redis"
	"github.com/gin-gonic/gin"
	"log"
)

func CreateRouter() (*gin.Engine, error) {
	r := gin.Default()

	// TODO: Change secret to .env variable
	store, err := redis.NewStore(10, "tcp", "localhost:6379", "", []byte("secret"))
	if err != nil {
		log.Println("Failed to create redis store")
		return nil, err
	}

	r.Use(sessions.Sessions("logins", store))

	r.LoadHTMLGlob("./public/*")

	r.GET("/", index)
	r.GET("/login", login)
	r.POST("/login", login)
	r.GET("/register", register)
	r.GET("/dashboard", dashboard)

	api := r.Group("/api")
	api.GET("/login", apiLogin)
	api.POST("/register", apiRegister)
	api.GET("/stream/:id", stream)
	api.GET("/capture/:id", capture)

	return r, nil
}
