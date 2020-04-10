package main

import "time"

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
