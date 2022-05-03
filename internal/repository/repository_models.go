package repository

import "time"

type ReportsRow struct {
	ID                  int
	StartedAt           time.Time
	CompletedAt         time.Time
	UsersToCreateCount  int
	UsersToUpdateCount  int
	UsersToDeleteCount  int
	GroupsToCreateCount int
	GroupsToUpdateCount int
	GroupsToDeleteCount int
}

type ErrorsRow struct {
	ID           int
	Err          string
	OccurredTime time.Time
}
