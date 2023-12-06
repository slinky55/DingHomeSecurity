package main

import (
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/hashicorp/mdns"
	"log"
	"net"
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
const DefaultDeviceDataPath = "data/devices"
const DefaultSaveFps = 1

var router *gin.Engine

var devices []Device

// get local ip
var stationIp string

func startDiscoveryService() {
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

			baseURL := "http://" + device.Ip + "/link"

			finalURL := baseURL + "?id=" + strconv.Itoa(int(device.ID)) + "&ip=" + stationIp

			get, err := http.Get(finalURL)
			if err != nil {
				log.Println("Failed to link to device at " + device.Ip)
				log.Println(get.Body)
				continue
			}

			if get.StatusCode != http.StatusOK {
				log.Println("Failed to link to device at " + device.Ip)
				log.Println(get.Body)
				continue
			}

			createDeviceDataFolder(device.ID)
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

func startHistoryService() {
	idCh := make(chan uint, 10)
	go func() {
		for id := range idCh {
			f, err := getFrame(id, "history")
			if err != nil {
				continue
			}

			fileName := fmt.Sprintf("%d.jpg", f.Timestamp.Unix())
			filePath := filepath.Join(f.FolderPath, fileName)

			err = os.WriteFile(filePath, f.Data, 0644)
			if err != nil {
				continue
			}
		}
	}()

	for {
		for _, device := range devices {
			idCh <- device.ID
		}
		time.Sleep(time.Second / DefaultSaveFps)
	}
}

var notifs = make(chan int, 10)

func startNotificationListener() {
	for id := range notifs {
		if id > len(devices) {
			continue
		}

		frame, err := getFrame(uint(id), "captures")
		if err != nil {
			log.Println("Failed to get notification capture")
			log.Println(err.Error())
			continue
		}

		fileName := fmt.Sprintf("%d.jpg", frame.Timestamp.Unix())
		filePath := filepath.Join(frame.FolderPath, fileName)

		err = os.WriteFile(filePath, frame.Data, 0644)
		if err != nil {
			log.Println("Failed to save notification capture")
		}
	}
}

func startCheckerService() {
	for {
		for _, device := range devices {
			baseURL := "http://" + device.Ip + "/link"
			finalURL := baseURL + "?id=" + strconv.Itoa(int(device.ID)) + "&ip=" + stationIp

			resp, err := http.Get(finalURL)
			if err != nil {
				log.Println("Failed to get device status")
				continue
			}

			if resp.StatusCode != http.StatusOK {
				log.Println("Failed to get device status")
				continue
			}
		}
		time.Sleep(time.Second * 60)
	}
}

func main() {
	args := os.Args

	host, _ := os.Hostname()
	addrs, _ := net.LookupIP(host)
	for _, addr := range addrs {
		if ipv4 := addr.To4(); ipv4 != nil {
			fmt.Println("Station ip: ", ipv4)
			stationIp = ipv4.String()
		}
	}

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
	go startDiscoveryService()

	if _, err := os.Stat(DefaultDeviceDataPath); os.IsNotExist(err) {
		err := os.Mkdir(DefaultDeviceDataPath, 0755)
		if err != nil {
			log.Println("Failed to create history folder")
			os.Exit(-1)
		}
	}

	for _, device := range devices {
		createDeviceDataFolder(device.ID)
	}

	go startHistoryService()

	go startNotificationListener()

	go startCheckerService()

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

func createDeviceDataFolder(id uint) {
	devicePath := path.Join(DefaultDeviceDataPath, strconv.Itoa(int(id)))

	stillsPath := path.Join(devicePath, "history")
	capturesPath := path.Join(devicePath, "captures")
	notifsPath := path.Join(devicePath, "notifs")

	if _, err := os.Stat(devicePath); os.IsNotExist(err) {
		err := os.Mkdir(devicePath, 0755)
		if err != nil {
			log.Println("Failed to create device history folder")
			log.Println(err.Error())
			return
		}

		err = os.Mkdir(stillsPath, 0755)
		if err != nil {
			log.Println("Failed to create device stills folder")
			log.Println(err.Error())
			return
		}

		err = os.Mkdir(capturesPath, 0755)
		if err != nil {
			log.Println("Failed to create device captures folder")
			log.Println(err.Error())
			return
		}

		err = os.Mkdir(notifsPath, 0755)
		if err != nil {
			log.Println("Failed to create device notifs folder")
			log.Println(err.Error())
			return
		}
	}
}
