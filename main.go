package main

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"runtime/debug"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/jlcampbell1991/campbell-radio-satellite-go/internal/clients"
	"github.com/jlcampbell1991/campbell-radio-satellite-go/internal/player"
	"github.com/jlcampbell1991/campbell-radio-satellite-go/internal/routes"
	"github.com/jlcampbell1991/campbell-radio-satellite-go/internal/utilities"
	"github.com/joho/godotenv"
)

func main() {
	defer func() {
		if r := recover(); r != nil {
			utilities.SaveCrashReport(fmt.Sprintf(
				"PANIC: %v\n\n%s",
				r,
				debug.Stack(),
			))
			os.Exit(1)
		}
	}()

	if err := godotenv.Load(); err != nil {
		log.Printf("No .env file found: %v", err)
	}
	port := os.Getenv("PORT")
	ffplay := os.Getenv("FFPLAY")
	amixer := os.Getenv("AMIXER")
	amixerDevice := os.Getenv("AMIXER_DEVICE")
	deviceName := os.Getenv("DEVICE_NAME")
	deviceId, err := uuid.Parse(os.Getenv("DEVICE_ID"))
	if err != nil {
		log.Fatalf("deviceId parsing error: %v", err)
	}

	httpClient := http.Client{}

	utilities.Go(func() {
		crashReportClient := clients.NewCrashReportClient(httpClient, deviceName, deviceId)
		utilities.ReadAndClearCrashReports(crashReportClient.PostReport)
	})

	player := player.NewPlayer(ffplay, amixer, amixerDevice)
	logger := clients.NewLogger(httpClient, deviceId)

	playerRoutes := routes.NewPlayerRoutes(httpClient, player, logger)

	r := chi.NewRouter()

	r.Mount("/v1", playerRoutes)

	http.ListenAndServe(":"+port, r)

	fmt.Println("Hi, this is the campbell-radio-satellte")
}
