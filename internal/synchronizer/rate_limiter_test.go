package synchronizer

import (
	"scim-integrations/internal/flags"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

const mockLimit = 1
const mockGreaterLimit = 2

func TestRateLimiter(t *testing.T) {
	t.Run("should reset when increment counter to the limit", func(t *testing.T) {
		*flags.RateLimiterFlag = true
		mock := newMockRateLimiter(mockLimit, time.Now())
		mock.Start()
		mock.IncreaseCounter()
		assert.Equal(t, mock.GetCounter(), 1)
		mock.VerifyLimit()
		assert.Equal(t, mock.GetCounter(), 0)
	})

	t.Run("should not reset when increment once", func(t *testing.T) {
		*flags.RateLimiterFlag = true
		mock := newMockRateLimiter(mockGreaterLimit, time.Now())
		mock.Start()
		mock.IncreaseCounter()
		assert.Equal(t, mock.GetCounter(), 1)
		mock.VerifyLimit()
		assert.Equal(t, mock.GetCounter(), 1)
	})

	t.Run("should reset when the time exceeds", func(t *testing.T) {
		*flags.RateLimiterFlag = true
		mockTime := time.Now().Add(time.Second * -30)
		mock := newMockRateLimiter(mockGreaterLimit, mockTime)
		mock.Start()
		mock.IncreaseCounter()
		assert.Equal(t, mock.GetCounter(), 1)
		mock.VerifyLimit()
		assert.Equal(t, mock.GetCounter(), 0)
	})
}

func newMockRateLimiter(limit int, startTime time.Time) RateLimiter {
	return &rateLimiterImpl{
		limit:     limit,
		rest:      func(time.Duration) {},
		startTime: startTime,
	}
}
