package synchronizer

import (
	"scim-integrations/internal/flags"
	"time"
)

const chokingWaitSeconds = 1
const rateLimitTime = time.Second * 30

type RateLimiter interface {
	Started() bool
	Start()
	IncreaseCounter()
	VerifyLimit()
	GetCounter() int
}

type rateLimiterImpl struct {
	limit     int
	counter   int
	startTime time.Time
	rest      func(time.Duration)
}

func newRateLimiter() RateLimiter {
	return &rateLimiterImpl{
		limit: 1000,
		rest:  rest,
	}
}

func (r *rateLimiterImpl) Started() bool {
	return !r.startTime.IsZero()
}

func (r *rateLimiterImpl) Start() {
	if r.startTime.IsZero() && *flags.RateLimiterFlag {
		r.startTime = time.Now()
	}
}

func (r *rateLimiterImpl) IncreaseCounter() {
	r.counter++
}

func (r *rateLimiterImpl) GetCounter() int {
	return r.counter
}

func (r *rateLimiterImpl) reset() {
	r.counter = 0
	r.startTime = time.Now()
}

func rest(secondsDiff time.Duration) {
	waitTime := time.Second * time.Duration(int(secondsDiff)+int(chokingWaitSeconds))
	time.Sleep(waitTime)
}

// VerifyLimit verifies if the request count reached the limit. If yes, it'll trigger a wait until pass the requests
// rate time.
func (r *rateLimiterImpl) VerifyLimit() {
	if !*flags.RateLimiterFlag {
		return
	}
	secondsDiff := time.Now().Sub(r.startTime)
	reachedLimit := r.counter >= r.limit
	if int(secondsDiff.Seconds()) < int(rateLimitTime.Seconds()) && !reachedLimit {
		return
	}
	if reachedLimit {
		r.rest(secondsDiff)
	}
	r.reset()
}
