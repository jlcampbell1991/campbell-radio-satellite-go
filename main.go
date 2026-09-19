package main

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"runtime/debug"

	"github.com/jlcampbell1991/campbell-radio-satellite-go/internal/clients"
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

	httpClient := http.Client{}

	utilities.Go(func() {
		crashReportClient := clients.NewCrashReportClient(httpClient)
		utilities.ReadAndClearCrashReports(crashReportClient.PostReport))
	})

	fmt.Println("Hi, this is the campbell-radio-satellte")
}
