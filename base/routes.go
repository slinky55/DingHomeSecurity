package main

import (
	"github.com/alexedwards/argon2id"
	"github.com/gin-contrib/sessions"
	"github.com/gin-gonic/gin"
	"net/http"
	"net/http/httputil"
	"net/url"
	"strconv"
)

func index(c *gin.Context) {
	sid := sessions.Default(c).Get("id")
	if sid != nil { // user is logged in
		res := dbConn.Model(&User{}).First(nil, sid.(uint))
		if res.RowsAffected > 0 { // user exists
			c.Redirect(http.StatusTemporaryRedirect, "/dashboard")
			return
		}
		sessions.Default(c).Clear()
	}

	c.Redirect(http.StatusTemporaryRedirect, "/login")
}

func stream(c *gin.Context) {
	var deviceUrl string

	if len(devices) == 0 {
		c.AbortWithStatus(http.StatusServiceUnavailable)
		return
	}

	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{
			"error": err.Error(),
		})
	}

	deviceUrl = "http://" + devices[id].Ip + "/stream"

	finalUrl, err := url.Parse(deviceUrl)
	if err != nil {
		c.JSON(http.StatusInternalServerError, err)
		return
	}

	proxy := httputil.NewSingleHostReverseProxy(finalUrl)
	proxy.ServeHTTP(c.Writer, c.Request)
}

func login(c *gin.Context) {
	sid := sessions.Default(c).Get("id")
	if sid != nil { // user is logged in
		res := dbConn.Model(&User{}).First(nil, sid.(uint))
		if res.RowsAffected > 0 { // user exists
			c.Redirect(http.StatusTemporaryRedirect, "/dashboard")
			return
		}
		sessions.Default(c).Clear()
	}

	c.HTML(http.StatusOK, "login.html", gin.H{})
}

func register(c *gin.Context) {
	sid := sessions.Default(c).Get("id")
	if sid != nil { // user is logged in
		res := dbConn.Model(&User{}).First(nil, sid.(uint))
		if res.RowsAffected > 0 { // user exists
			c.Redirect(http.StatusTemporaryRedirect, "/dashboard")
			return
		}
		sessions.Default(c).Clear()
	}

	c.HTML(http.StatusOK, "register.html", gin.H{})
}

func dashboard(c *gin.Context) {
	sid := sessions.Default(c).Get("id")
	if sid == nil { // not logged in
		c.Redirect(http.StatusTemporaryRedirect, "/login")
		return
	}

	res := dbConn.Model(&User{}).First(nil, sid.(uint))
	if res.RowsAffected == 0 { // logged in but user does not exist
		sessions.Default(c).Clear()
		c.Redirect(http.StatusTemporaryRedirect, "/login")
		return
	}

	type Data struct {
		Ids []uint
	}

	var data Data
	for i := 0; i < len(devices); i++ {
		data.Ids = append(data.Ids, devices[i].ID-1)
	}

	c.HTML(http.StatusOK, "dashboard.html", gin.H{
		"Data": data,
	})
}

func apiLogin(c *gin.Context) {
	session := sessions.Default(c)

	type Login struct {
		Username string `form:"username"`
		Password string `form:"password"`
	}
	var login Login

	err := c.ShouldBind(&login)
	if err != nil {
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{
			"error": err.Error(),
		})
		return
	}

	var user User
	res := dbConn.First(&user, "username = ?", login.Username)
	if res.RowsAffected == 0 { // No user found with username
		c.AbortWithStatusJSON(http.StatusConflict, gin.H{
			"error": "incorrect username or password",
		})
		return
	}

	match, _, err := argon2id.CheckHash(login.Password, user.PasswordHash)
	if err != nil {
		c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{
			"error": err.Error(),
		})
		return
	}

	if !match {
		c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
			"error": "incorrect username or password",
		})
		return
	}

	// We should have correct username and password here
	session.Set("id", user.ID)
	err = session.Save()
	if err != nil {
		c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{
			"error": err.Error(),
		})
		return
	}

	c.Redirect(http.StatusTemporaryRedirect, "/dashboard")
}

func apiRegister(c *gin.Context) {
	type Registration struct {
		Username string `form:"username" binding:"required"`
		Password string `form:"password" binding:"required"`
	}

	var reg Registration

	err := c.ShouldBind(&reg)
	if err != nil {
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{
			"error": err.Error(),
		})
		return
	}

	res := dbConn.Model(&User{}).Find(nil, "username = ?", reg.Username)

	if res.RowsAffected > 0 {
		c.AbortWithStatusJSON(http.StatusConflict, gin.H{
			"error": "username is taken",
		})
		return
	}

	hash, err := argon2id.CreateHash(reg.Password, argon2id.DefaultParams)
	if err != nil {
		c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{
			"error": err.Error(),
		})
		return
	}

	isAdmin := false

	res = dbConn.Model(&User{}).Find(nil)
	if res.RowsAffected == 0 {
		isAdmin = true
	}

	dbConn.Create(&User{
		Username:     reg.Username,
		PasswordHash: hash,
		IsAdmin:      isAdmin,
	})

	c.Redirect(http.StatusTemporaryRedirect, "/login")
}

func notify(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{
			"error": err.Error(),
		})
		return
	}

	notifs <- id

	c.Status(http.StatusOK)
}
