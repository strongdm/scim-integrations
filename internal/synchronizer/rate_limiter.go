package synchronizer

import (
	"time"
)

const chokingWaitSeconds = 1
const rateLimitTime = time.Second * 30

type RateLimiter struct {
	limit     int
	index     int
	startTime time.Time
}

func NewRateLimiter() *RateLimiter {
	return &RateLimiter{
		// limit: 1000,
		limit: 1,
	}
}

func (r *RateLimiter) Start() {
	if r.startTime.IsZero() {
		r.startTime = time.Now()
	}
}

func (r *RateLimiter) IncreaseIdx() {
	r.index++
}

func (r *RateLimiter) reset() {
	r.index = 0
	r.startTime = time.Now()
}

func (r *RateLimiter) VerifyLimit() {
	secondsDiff := time.Now().Sub(r.startTime).Seconds()
	reachedLimit := r.index >= r.limit
	if secondsDiff < rateLimitTime.Seconds() && !reachedLimit {
		return
	}
	waitTime := time.Second * time.Duration(int(secondsDiff)+int(chokingWaitSeconds))
	if reachedLimit {
		time.Sleep(waitTime)
	}
	r.reset()
}
