package clients

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"net/http"

	"github.com/google/uuid"
)

type Logger interface {
	Debug(message string) error
	Info(message string) error
	Warn(message string) error
	Error(message string) error
}

type logger struct {
	client   http.Client
	playerID uuid.UUID
	host     string
}

func NewLogger(client http.Client, playerID uuid.UUID) Logger {
	return &logger{
		client:   client,
		playerID: playerID,
		host:     "192.168.50.185:9090",
	}
}

func (l *logger) Debug(message string) error {
	return l.makeReq(message, "debug")
}

func (l *logger) Info(message string) error {
	return l.makeReq(message, "info")
}

func (l *logger) Warn(message string) error {
	return l.makeReq(message, "warn")
}

func (l *logger) Error(message string) error {
	return l.makeReq(message, "error")
}

func (l *logger) makeReq(message string, level string) error {
	url := fmt.Sprintf("http://%s/logs", l.host)

	data := struct {
		PlayerID uuid.UUID `json:"player_id"`
		Message  string    `json:"message"`
		Level    string    `json:"level"`
	}{
		PlayerID: l.playerID,
		Message:  message,
		Level:    level,
	}

	body, err := json.Marshal(data)
	if err != nil {
		return fmt.Errorf("failed to marshal log request: %w", err)
	}

	log.Printf("Sending log request to remote host %s", url)
	log.Printf("%s", body)

	req, err := http.NewRequest(
		http.MethodPost,
		url,
		bytes.NewReader(body),
	)
	if err != nil {
		return fmt.Errorf("failed to create log request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")

	resp, err := l.client.Do(req)
	if err != nil {
		log.Printf("Failed to send log request to remote host %s: %v", url, err)
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		err := fmt.Errorf("remote logger returned status %s", resp.Status)
		log.Printf("Failed to send log request to remote host %s: %v", url, err)
		return err
	}

	return nil
}
