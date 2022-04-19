package synchronizer

import (
	"errors"
	"fmt"
	"os"
	"strings"

	"github.com/cenkalti/backoff/v4"
)

type EntityScope int

type Retrier interface {
	Run(fn func() error, actionDescription string) error
	GetBackoffConfig() backoff.BackOff
	IncreaseCounter()
	GetCounter() int
	ExceededLimit() bool
	GetRateLimiter() RateLimiter
	setEntityScope(EntityScope)
	getEntityScope() EntityScope
	ErrorIsUnexpected(err error) bool
}

const (
	UserScope EntityScope = iota
	GroupScope

	retryLimitCount = 4

	alreadyExistsErrMsg  = "values are already in use"
	wrongMemberIDErrMsg  = "cannot parse member id"
	memberNotFoundErrMsg = "not found"
)

var mappedScopeErrs = map[EntityScope][]string{
	UserScope: {
		alreadyExistsErrMsg,
	},
	GroupScope: {
		wrongMemberIDErrMsg,
		memberNotFoundErrMsg,
	},
}

type retrierImpl struct {
	rateLimiter RateLimiter
	counter     int
	limit       int
	try         func(retrier Retrier, fn func() error, actionDescription string) func() error
	scope       EntityScope
}

func newRetrier(rateLimiter RateLimiter) Retrier {
	return &retrierImpl{
		rateLimiter: rateLimiter,
		counter:     0,
		limit:       retryLimitCount,
		try:         try,
	}
}

func (r *retrierImpl) IncreaseCounter() {
	r.counter++
}

func (r *retrierImpl) GetCounter() int {
	return r.counter
}

func (r *retrierImpl) ExceededLimit() bool {
	return r.counter >= r.limit
}

func (r *retrierImpl) GetRateLimiter() RateLimiter {
	return r.rateLimiter
}

func (r *retrierImpl) setEntityScope(scope EntityScope) {
	r.scope = scope
}

func (r *retrierImpl) getEntityScope() EntityScope {
	return r.scope
}

func (r *retrierImpl) GetBackoffConfig() backoff.BackOff {
	return backoff.WithMaxRetries(backoff.NewExponentialBackOff(), uint64(r.limit))
}

func (r *retrierImpl) Run(fn func() error, actionDescription string) error {
	err := backoff.Retry(r.try(r, fn, actionDescription), r.GetBackoffConfig())
	return err
}

func (r *retrierImpl) ErrorIsUnexpected(err error) bool {
	errs := mappedScopeErrs[r.scope]
	for _, msg := range errs {
		if strings.Contains(err.Error(), msg) {
			return false
		}
	}
	return true
}

func try(retrier Retrier, fn func() error, actionDescription string) func() error {
	return func() error {
		limiter := retrier.GetRateLimiter()
		limiter.VerifyLimit()
		err := fn()
		limiter.IncreaseCounter()
		if err != nil {
			if !retrier.ErrorIsUnexpected(err) {
				fmt.Fprintln(os.Stderr, err)
				return nil
			}
			retrier.IncreaseCounter()
			if retrier.ExceededLimit() {
				return errors.New("retry limit exceeded with the following error: " + err.Error())
			}
			fmt.Fprintf(os.Stderr, "Failed %s. Retrying the operation for the %dst time\n", actionDescription, retrier.GetCounter())
			return err
		}
		return nil
	}
}
