package source

import (
	"context"
)

const FETCH_PAGE_SIZE = 500

type BaseSource interface {
	FetchUsers(ctx context.Context) ([]User, error)
	ExtractGroupsFromUsers([]User) []UserGroup
}

func ByFlag(name string) BaseSource {
	if name == "google" {
		return NewGoogleSource()
	}
	return nil
}
