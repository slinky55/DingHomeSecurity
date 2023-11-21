package main

import (
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

var dbConn *gorm.DB

func initDB(path string) error {
	var err error
	dbConn, err = gorm.Open(sqlite.Open(path), &gorm.Config{})
	if err != nil {
		return err
	}

	err = dbConn.AutoMigrate(&User{}, &Device{})
	if err != nil {
		return err
	}

	return nil
}
