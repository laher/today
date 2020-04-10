package main

import (
	"os"
	"path/filepath"
	"time"
)

func getBaseDir() (string, error) {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(homeDir, "today"), nil
}

func getTodayFilename() (string, error) {
	base, err := getBaseDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(base, "today.md"), nil
}

func getRecurringFilename() (string, error) {
	base, err := getBaseDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(base, "recurring.today.md"), nil
}

func getArchiveFilename(forTime time.Time) (string, error) {
	base, err := getBaseDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(base, forTime.Format(filepath.Join("2006", "01", "02"))+".today.md"), nil
}
