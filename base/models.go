package main

import "gorm.io/gorm"

type User struct {
	gorm.Model
	Username     string
	PasswordHash string
	IsAdmin      bool
}

type Device struct {
	gorm.Model
	Ip         string
	Hostname   string
	Name       string
	IsDoorbell bool
}
