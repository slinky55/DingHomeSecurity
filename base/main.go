package main

import (
	"github.com/gin-gonic/gin"
	"github.com/hashicorp/mdns"
	"log"
	"os"
	"strings"
	"time"
)

const DefaultPort string = "8080"

var router *gin.Engine

var devices []Device

func discoverDevices() {
	services := make(chan *mdns.ServiceEntry, 10)
	go func() {
		for service := range services {
			res := dbConn.Model(&Device{}).First(nil, "ip = ?", service.AddrV4.String())

			if res.RowsAffected > 0 {
				continue
			}

			if !strings.Contains(service.Name, "dinghs") {
				continue
			}

			device := Device{
				Ip:       service.AddrV4.String(),
				Hostname: service.Name,
			}

			dbConn.Create(&device)
			devices = append(devices, device)
		}
	}()

	for {
		params := mdns.DefaultParams("_dinghs._tcp")
		params.DisableIPv6 = true
		params.Entries = services
		err := mdns.Query(params)
		if err != nil {
			log.Println(err)
		}
		time.Sleep(time.Second * 5)
	}
}

func main() {
	args := os.Args

	err := initDB("./db.sqlite")
	if err != nil {
		log.Println("Failed to connect to database")
		log.Println(err.Error())
		return
	}

	dbConn.Find(&devices)
	go discoverDevices()

	router, err = CreateRouter()
	if err != nil {
		log.Println(err)
		return
	}

	var port string

	if len(args) > 1 {
		port = args[1]
	} else {
		port = DefaultPort
	}

	err = router.Run(":" + port)
	if err != nil {
		log.Fatalln(err)
	}
}
