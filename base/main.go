package main

import (
	"github.com/gin-gonic/gin"
	"github.com/gin-contrib/sessions"
	"github.com/gin-contrib/sessions/redis"
	"github.com/joho/godotenv"
	"log"
	"net/http"
	"github.com/jmoiron/sqlx"
	"github.com/go-sql-driver/mysql
)

func main() {
	log.Println("Starting server...")

	err := godotenv.Load()
	if err != nil {
		log.Fatalln("Error loading .env file")
	}

	dbInit(true)
	router := gin.Default()

	store, err :=
		redis.NewStore(10, "tcp", "localhost:6379", "", []byte("secret"))

	if err != nil {
		log.Fatal(err)
	}

	store.Options(sessions.Options{
		Path:     "/",
		MaxAge:   60 * 60,
		HttpOnly: true,
		Secure:   false,
		SameSite: http.SameSiteDefaultMode,
	})

	router.Use(sessions.Sessions("session", store))

	router.LoadHTMLFiles("public/index.html")

	router.Static("/images", "./images")

	router.GET("/ping", func(c *gin.Context) {
		c.IndentedJSON(http.StatusOK, gin.H{
			"message": "pong",
		})
	})
	
	router.GET("/login", func(c *gin.Context) {
		c.HTML(http.StatusOK, "index.html", gin.H{})
	})

	router.POST("/login", func(c *gin.Context) {
		username := c.PostForm("username")
		password := c.PostForm("password")
		if authenticateUser(username, password) {
			session := sessions.Default(c)
			session.Set("username", username)
			session.Save()

			c.Redirect(http.StatusSeeOther, "/welcome")
		} else {
			c.String(http.StatusUnauthorized, "Invalid credentials")
		}
	})
	
	err = router.Run(":7100")
	if err != nil {
		log.Fatal(err)
	}
}
func authenticateUser(username, password string) bool {
	storedPasswordHash, err := getHashedPassword(username)

	if err != nil {
		fmt.Println("Username not found")
		return false
	}
	err = bcrypt.CompareHashAndPassword([]byte(storedPasswordHash), []byte(password))
	return err == nil
}
