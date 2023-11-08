package main

import (
	"net/http"
	"net/http/httputil"
	"net/url"

	"github.com/gin-gonic/gin"
)

func proxyStream(c *gin.Context) {
    // ESP32 internal address
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
	router := gin.Default()

	router.LoadHTMLFiles("./public/index.html")

	router.GET("/", func(c *gin.Context) {
		c.HTML(http.StatusOK, "index.html", nil)
	})
	router.GET("/stream", proxyStream)

	// Run the server
	router.Run(":80") // Change this port to match your server configuration
}
