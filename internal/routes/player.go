package routes

import (
	"encoding/json"
	"log"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/jlcampbell1991/campbell-radio-satellite-go/internal/clients"
	"github.com/jlcampbell1991/campbell-radio-satellite-go/internal/models"
	"github.com/jlcampbell1991/campbell-radio-satellite-go/internal/player"
)

type playerRoutes struct {
	client http.Client
	player player.Player
	logger clients.Logger
}

func NewPlayerRoutes(
	client http.Client,
	player player.Player,
	logger clients.Logger,
) chi.Router {
	r := chi.NewRouter()

	p := &playerRoutes{
		client: client,
		player: player,
		logger: logger,
	}

	r.Put("/play", p.Play)
	r.Put("/pause-resume", p.PauseResume)
	r.Put("/volume", p.SetVolume)
	r.Get("/volume", p.GetVolume)

	return r
}

type playRequest struct {
	Medium   models.Media `json:"medium"`
	Callback string       `json:"callback"`
}
type volumeRequest struct {
	Volume int `json:"volume"`
}

func callbackFn(client http.Client, callbackURL string) error {
	log.Printf("Sending PUT request to '%s'", callbackURL)

	req, err := http.NewRequest(http.MethodPut, callbackURL, nil)
	if err != nil {
		return err
	}

	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	return nil
}

func (p *playerRoutes) Play(w http.ResponseWriter, r *http.Request) {
	var req playRequest

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	if err := p.player.Play(
		req.Medium,
		func() {
			err := callbackFn(p.client, req.Callback)
			if err != nil {
				p.logger.Error("error while making callback on Play: %d", err)
			}
		},
	); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
}

func (p *playerRoutes) PauseResume(w http.ResponseWriter, r *http.Request) {
	if err := p.player.PauseResume(); err != nil {
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}

	w.WriteHeader(http.StatusOK)
}

func (p *playerRoutes) SetVolume(w http.ResponseWriter, r *http.Request) {
	var req volumeRequest

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	if err := p.player.SetVolume(req.Volume); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
}

func (p *playerRoutes) GetVolume(w http.ResponseWriter, r *http.Request) {
	volume, err := p.player.GetVolume()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(volume)
}
