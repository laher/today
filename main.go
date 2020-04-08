package main

import (
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"time"
)

func main() {
	args := os.Args[1:]

	if len(args) < 1 {
		fmt.Fprintln(os.Stderr, "Please specify a subcommand")
		os.Exit(1)
	}
	var err error
	switch args[0] {
	case "rollover":
		err = rollover(args)
	case "days":
		err = days(args)
	default:
		err = errors.New("Unrecognised subcommand")
	}
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error handling [%s]: %v\n", args[0], err)
		os.Exit(1)
	}
}

type task struct {
	Description string
	Status      string
	Tags        []string
	Created     time.Time
	Updated     time.Time
	Completed   time.Time

	Subtasks []tasks

	RecurType recurType
	From      time.Time
	Until     time.Time
}

type recurType string

const (
	custom   recurType = "custom"
	hourly   recurType = "hourly"
	daily    recurType = "daily"
	weekly   recurType = "weekly"
	weekdays recurType = "weekdays"
)

type tasks struct {
	Tasks []task
}

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

func loadToday() (tasks, error) {
	file, err := getTodayFilename()
	if err != nil {
		return tasks{}, err
	}
	return parseFile(file)
}

func parseFile(file string) (tasks, error) {
	f, err := os.Open(file)
	if err != nil {
		return tasks{}, err
	}
	return parse(f)
}

func parse(r io.Reader) (tasks, error) {
	// TODO!
	return tasks{}, nil
}

func loadRecurring() (tasks, error) {
	file, err := getRecurringFilename()
	if err != nil {
		return tasks{}, err
	}
	return parseFile(file)
}

func archiveToday() error {
	return nil
}

func newToday(current tasks, recurring tasks) error {
	return nil
}

func rollover(args []string) error {
	// load today
	today, err := loadToday()
	if err != nil {
		return err
	}
	// load recurring
	recurring, err := loadRecurring()
	if err != nil {
		return err
	}
	if err := archiveToday(); err != nil {
		return err
	}
	return newToday(today, recurring)
}

func days(args []string) error {
	// print some days
	for i := 0; i < 5; i++ {
		fmt.Println(time.Now().Add(time.Duration(i) * time.Hour * 24).Format("2006-01-02, Mon"))
	}
	return nil
}
