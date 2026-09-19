package utilities

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"runtime/debug"
	"strings"
	"time"
)

func SaveCrashReport(report string) {
	exe, err := os.Executable()
	if err != nil {
		log.Printf("failed to get executable path: %v", err)
		return
	}

	dir := filepath.Dir(exe)

	filename := fmt.Sprintf(
		"campbell-radio-crash-%s.log",
		time.Now().Format("20060102-150405"),
	)

	log.Printf("filename: %v", filepath.Join(dir, filename))

	if err := os.WriteFile(
		filepath.Join(dir, filename),
		[]byte(report),
		0644,
	); err != nil {
		log.Printf("failed to write crash report: %v", err)
	}
}

func Go(fn func()) {
	go func() {
		defer func() {
			if r := recover(); r != nil {
				SaveCrashReport(fmt.Sprintf(
					"PANIC: %v\n\n%s",
					r,
					debug.Stack(),
				))
				os.Exit(1)
			}
		}()
		fn()
	}()
}

func ReadAndClearCrashReports(saveCrashReport func(string, time.Time) error) error {
	exe, err := os.Executable()
	if err != nil {
		return err
	}

	dir := filepath.Dir(exe)

	entries, err := os.ReadDir(dir)
	if err != nil {
		return fmt.Errorf("reading crash report directory: %w", err)
	}

	for _, entry := range entries {
		name := entry.Name()

		if entry.IsDir() ||
			!strings.HasPrefix(name, "campbell-radio-crash-") ||
			!strings.HasSuffix(name, ".log") {
			continue
		}

		path := filepath.Join(dir, name)

		data, err := os.ReadFile(path)
		if err != nil {
			return fmt.Errorf("reading crash report %q: %w", name, err)
		}

		report := string(data)

		timestamp := strings.TrimSuffix(
			strings.TrimPrefix(name, "campbell-radio-crash-"),
			".log",
		)

		createdAt, err := time.Parse("20060102-150405", timestamp)
		if err != nil {
			return fmt.Errorf("parsing crash report date %q: %w", name, err)
		}

		err = saveCrashReport(report, createdAt)
		if err != nil {
			return err
		}

		if err := os.Remove(path); err != nil {
			return fmt.Errorf("removing crash report %q: %w", name, err)
		}
	}

	return nil
}
