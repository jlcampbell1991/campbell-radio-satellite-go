package clients

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net"
	"net/http"
	"os"
	"time"
)

type CrashReportClient interface {
	PostReport(report string, createdAt time.Time) error
	// PostDevice(deviceName, ip string, deviceId uuid.UUID) error
}

type crashReportClient struct {
	client http.Client
}

func NewCrashReportClient(client http.Client) CrashReportClient {
	return &crashReportClient{client: client}
}

type crashReportRequest struct {
	Report     string    `json:"report"`
	CreatedAt  time.Time `json:"created_at"`
	DeviceName string    `json:"device_name"`
	Ip         string    `json:"ip"`
}

func getLocalIP() (string, error) {
	conn, err := net.Dial("udp", "8.8.8.8:80")
	if err != nil {
		return "", err
	}
	defer conn.Close()

	localAddr, ok := conn.LocalAddr().(*net.UDPAddr)
	if !ok {
		return "", fmt.Errorf("unexpected local address type: %T", conn.LocalAddr())
	}

	return localAddr.IP.String(), nil
}

func (c *crashReportClient) PostReport(report string, createdAt time.Time) error {
	ip, err := getLocalIP()
	if err != nil {
		return err
	}

	deviceId := os.Getenv("DEVICE_ID")
	deviceName := os.Getenv("DEVICE_NAME")

	url := fmt.Sprintf("http://192.168.50.185:9090/devices/%v/crash-reports", deviceId)

	data, err := json.Marshal(crashReportRequest{
		Report:     report,
		CreatedAt:  createdAt,
		DeviceName: deviceName,
		Ip:         ip,
	})
	if err != nil {
		return err
	}

	reader := bytes.NewReader(data)

	req, err := http.NewRequest(http.MethodPost, url, reader)
	if err != nil {
		return err
	}

	_, err = c.client.Do(req)

	return err
}

// type device struct {
// 	Name string    `json:"name"`
// 	ID   uuid.UUID `json:"id"`
// 	Ip   string    `json:"ip"`
// }

// func (c *devicesClient) PostDevice(deviceName, ip string, deviceId uuid.UUID) error {
// 	path := fmt.Sprintf("/devices")

// 	data, err := json.Marshal(device{
// 		ID:   deviceId,
// 		Name: deviceName,
// 		Ip:   ip,
// 	})
// 	if err != nil {
// 		return err
// 	}

// 	_, err = RetryFetchFrom[string](
// 		c.client,
// 		&c.db,
// 		http.MethodPost,
// 		path,
// 		&data,
// 	)

// 	return err
// }
