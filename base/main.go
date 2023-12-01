package main

import (
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/hashicorp/mdns"
	"io"
	"log"
	"net/http"
	"os"
	"path"
	"path/filepath"
	"strconv"
	"strings"
	"time"
)

const DefaultPort string = "8080"
const DefaultDataFolder = "data"
const DefaultHistoryFolder = "data/history"
const DefaultSaveFps = 4

var router *gin.Engine

var devices []Device

func discoverDevices() {
	services := make(chan *mdns.ServiceEntry, 10)
	go func() {
		for service := range services {
			if !strings.Contains(service.Name, "dinghs") {
				continue
			}

			res := dbConn.Model(&Device{}).First(nil, "ip = ?", service.AddrV4.String())

			if res.RowsAffected > 0 {
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

type Frame struct {
	Data       []byte
	Timestamp  time.Time
	FolderPath string
}

func startHistoryService() {
	frameCh := make(chan Frame, 10)
	go func() {
		for f := range frameCh {
			fileName := fmt.Sprintf("%d.jpg", f.Timestamp.Unix())
			filePath := filepath.Join(f.FolderPath, fileName)

			err := os.WriteFile(filePath, f.Data, 0644)
			if err != nil {
				continue
			}
		}
	}()

	for {
		for _, device := range devices {
			resp, err := http.Get("http://" + device.Ip + "/capture")
			if err != nil {
				continue
			}

			data, err := io.ReadAll(resp.Body)
			if err != nil {
				continue
			}

			frame := Frame{
				Data:       data,
				Timestamp:  time.Now(),
				FolderPath: path.Join(DefaultHistoryFolder, strconv.Itoa(int(device.ID))),
			}

			frameCh <- frame

			err = resp.Body.Close()
			if err != nil {
				continue
			}
		}
		time.Sleep(time.Second / DefaultSaveFps)
	}
}

func main() {
	args := os.Args

	if _, err := os.Stat(DefaultDataFolder); os.IsNotExist(err) {
		err := os.Mkdir(DefaultDataFolder, 0755)
		if err != nil {
			log.Println("Failed to create data folder")
			os.Exit(-1)
		}
	}
	err := initDB("data/db.sqlite")
	if err != nil {
		log.Println("Failed to connect to database")
		log.Println(err.Error())
		os.Exit(-1)
	}

	dbConn.Find(&devices)
	go discoverDevices()

	if _, err := os.Stat(DefaultHistoryFolder); os.IsNotExist(err) {
		err := os.Mkdir(DefaultHistoryFolder, 0755)
		if err != nil {
			log.Println("Failed to create history folder")
			os.Exit(-1)
		}
	}

	for _, device := range devices {
		devicePath := path.Join(DefaultHistoryFolder, strconv.Itoa(int(device.ID)))
		if _, err := os.Stat(devicePath); os.IsNotExist(err) {
			err := os.Mkdir(devicePath, 0755)
			if err != nil {
				log.Println("Failed to create history folder")
				os.Exit(-1)
			}
		}
	}

	go startHistoryService()

	router, err = CreateRouter()
	if err != nil {
		log.Println(err)
		os.Exit(-1)
	}

	var port string

	if len(args) > 1 {
		port = args[1]
	} else {
		port = DefaultPort
	}

	err = router.Run(":" + port)
	if err != nil {
		log.Println("Failed to start router")
		log.Println(err)
		os.Exit(-1)
	}
}
