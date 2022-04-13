package synchronizer

import (
	"time"
)

const chokingWaitSeconds = 1
const rateLimitTime = time.Second * 30

type RateLimiter struct {
	limit     int
	counter   int
	startTime time.Time
}

func NewRateLimiter() *RateLimiter {
	return &RateLimiter{
		limit: 1000,
	}
}

func (r *RateLimiter) Start() {
	if r.startTime.IsZero() {
		r.startTime = time.Now()
	}
}

func (r *RateLimiter) IncreaseCounter() {
	r.counter++
}

func (r *RateLimiter) reset() {
	r.counter = 0
	r.startTime = time.Now()
}

func (r *RateLimiter) VerifyLimit() {
	secondsDiff := time.Now().Sub(r.startTime).Seconds()
	reachedLimit := r.counter >= r.limit
	if secondsDiff < rateLimitTime.Seconds() && !reachedLimit {
		return
	}
	if reachedLimit {
		waitTime := time.Second * time.Duration(int(secondsDiff)+int(chokingWaitSeconds))
		time.Sleep(waitTime)
	}
	r.reset()
}
