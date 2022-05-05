package repository

import "time"

type ReportsRow struct {
	ID                 int
	StartedAt          time.Time
	CompletedAt        time.Time
	CreatedUsersCount  int
	CreatedGroupsCount int
	UpdatedUsersCount  int
	UpdatedGroupsCount int
	DeletedUsersCount  int
	DeletedGroupsCount int
	Succeed            int
}
