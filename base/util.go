package main

import (
	"io"
	"net/http"
	"path"
	"strconv"
	"time"
)

func getFrame(id uint, folder string) (Frame, error) {
	resp, err := http.Get("http://" + devices[id-1].Ip + "/capture")
	if err != nil {
		return Frame{}, err
	}

	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return Frame{}, err
	}

	frame := Frame{
		Data:       data,
		Timestamp:  time.Now(),
		FolderPath: path.Join(DefaultDeviceDataPath, strconv.Itoa(int(id)), folder),
	}

	_ = resp.Body.Close()

	return frame, nil
}
