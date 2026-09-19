package models

import "github.com/google/uuid"

type Media struct {
	ID            uuid.UUID `json:"id"`
	Name          string    `json:"name"`
	Location      *string   `json:"location"`
	Active        bool      `json:"active"`
	LastPlayed    *string   `json:"last_played"`
	Plays         int       `json:"plays"`
	CreatedAt     string    `json:"created_at"`
	UpdatedAt     string    `json:"updated_at"`
	FileName      *string   `json:"file_name"`
	Artist        string    `json:"artist"`
	Album         string    `json:"album"`
	Duration      int       `json:"duration"`
	Loudness      float64   `json:"loudness"`
	LRange        float64   `json:"lrange"`
	Peak          float64   `json:"peak"`
	Gain          float64   `json:"gain"`
	MediaType     string    `json:"media_type"`
	Deleted       bool      `json:"deleted"`
	PlaybackStart int       `json:"playback_start"`
	PlaybackEnd   *int      `json:"playback_end"`
	FadeIn        int       `json:"fade_in"`
	FadeOut       int       `json:"fade_out"`
	Src           *string   `json:"src"`
	Track         *int      `json:"track"`
}
