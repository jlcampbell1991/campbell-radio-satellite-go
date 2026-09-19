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
	Debug(message string, args ...any) error
	Info(message string, args ...any) error
	Warn(message string, args ...any) error
	Error(message string, args ...any) error
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

func (l *logger) Debug(message string, args ...any) error {
	return l.makeReq("debug", message, args...)
}

func (l *logger) Info(message string, args ...any) error {
	return l.makeReq("info", message, args...)
}

func (l *logger) Warn(message string, args ...any) error {
	return l.makeReq("warn", message, args...)
}

func (l *logger) Error(message string, args ...any) error {
	return l.makeReq("error", message, args...)
}

func (l *logger) makeReq(level, message string, args ...any) error {
	msg := fmt.Sprintf(message, args)

	log.Printf("[%v] %v", level, msg)

	url := fmt.Sprintf("http://%s/logs", l.host)

	data := struct {
		PlayerID uuid.UUID `json:"player_id"`
		Message  string    `json:"message"`
		Level    string    `json:"level"`
	}{
		PlayerID: l.playerID,
		Message:  msg,
		Level:    level,
	}

	body, err := json.Marshal(data)
	if err != nil {
		return fmt.Errorf("failed to marshal log request: %w", err)
	}

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
