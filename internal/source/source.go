package source

import (
	"context"
	"net/http"
)

var HTTPClient *http.Client

const FETCH_PAGE_SIZE = 500

type BaseSource interface {
	FetchUsers(ctx context.Context) ([]SourceUser, error)
	ExtractGroupsFromUsers([]SourceUser) []SourceUserGroup
}

func ByFlag(name string) BaseSource {
	if name == "google" {
		return NewGoogleSource()
	}
	return nil
}
