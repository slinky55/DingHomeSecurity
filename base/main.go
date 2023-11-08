package main

import (
	"os"

	"net/http"
	"net/http/httputil"
	"net/url"

	"github.com/gin-gonic/gin"
)

var IPMap map[int]string

func proxyStream(c *gin.Context) {
    esp32StreamURL := "http://192.168.0.190/stream"
    
    url, err := url.Parse(esp32StreamURL)
    if err != nil {
	c.JSON(http.StatusInternalServerError, err)
	return
    }

    proxy := httputil.NewSingleHostReverseProxy(url)
    proxy.ServeHTTP(c.Writer, c.Request)
}

func main() {
	args := os.Args

	router := gin.Default()

	router.LoadHTMLFiles("./public/index.html")

	router.GET("/", func(c *gin.Context) {
		c.HTML(http.StatusOK, "index.html", nil)
	})

	router.GET("/stream", proxyStream)

	// Run the server
	if len(args) > 1 {
		router.Run(":" + args[1])
	} 

	router.Run(":8080")
}
