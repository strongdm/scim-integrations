package synchronizer

import (
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
	ErrorIsExpected(err error) bool
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
	return backoff.Retry(r.try(r, fn, actionDescription), r.GetBackoffConfig())
}

func (r *retrierImpl) ErrorIsExpected(err error) bool {
	errs := mappedScopeErrs[r.scope]
	for _, msg := range errs {
		if strings.Contains(err.Error(), msg) {
			return true
		}
	}
	return false
}

func try(retrier Retrier, fn func() error, actionDescription string) func() error {
	return func() error {
		if !retrier.GetRateLimiter().Started() {
			retrier.GetRateLimiter().Start()
		}
		limiter := retrier.GetRateLimiter()
		limiter.VerifyLimit()
		err := fn()
		limiter.IncreaseCounter()
		if err != nil {
			if !retrier.ErrorIsExpected(err) {
				return err
			}
			retrier.IncreaseCounter()
			if retrier.ExceededLimit() {
				fmt.Fprintf(os.Stderr, "retry limit exceeded with the following error: %s", err.Error())
			} else {
				fmt.Fprintf(os.Stderr, "Failed %s. Retrying the operation for the %dst time\n", actionDescription, retrier.GetCounter())
			}
			return nil
		}
		return nil
	}
}
