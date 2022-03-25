package idp

import (
	"context"
	"net/http"
)

var HTTPClient *http.Client

const FETCH_PAGE_SIZE = 500

type BaseIdP interface {
	FetchUsers(ctx context.Context) ([]IdPUser, error)
}

func ByFlag(name string) BaseIdP {
	if name == "google" {
		return NewGoogleIdP()
	}
	return nil
}
