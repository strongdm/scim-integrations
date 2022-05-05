package synchronizer

import (
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

var expectedErrMsg = alreadyExistsErrMsg
var unexpectedErrMsg = "random error message"

func TestRetrier(t *testing.T) {
	t.Run("should try once when return no errors", func(t *testing.T) {
		mock := newMockCounter()
		mock.On("count").Return(nil)
		retrier := newRetrier(newRateLimiter())

		err := retrier.Run(mock.count, "test")

		assert.Equal(t, 1, len(mock.Calls))
		assert.Nil(t, err)
	})

	t.Run("should try once when return an expected error", func(t *testing.T) {
		mock := newMockCounter()
		mock.On("count").Return(errors.New(expectedErrMsg))
		retrier := newRetrier(newRateLimiter())

		err := retrier.Run(mock.count, "test")

		assert.Equal(t, 1, len(mock.Calls))
		assert.Nil(t, err)
	})

	t.Run("should try twice when in the first time return an unexpected error and then return nil", func(t *testing.T) {
		mock := newMockCounter()
		retrier := newRetrier(getExceededRateLimiter())
		mock.On("count").Return(errors.New(unexpectedErrMsg)).Times(5)

		err := retrier.Run(mock.count, "test")

		assert.Equal(t, 5, len(mock.Calls))
		assert.NotNil(t, err)
	})

	t.Run("should try twice when in the first time return an unexpected error and then return nil", func(t *testing.T) {
		mock := newMockCounter()
		retrier := newRetrier(getExceededRateLimiter())
		mock.On("count").Return(errors.New(expectedErrMsg)).Once()
		mock.On("count").Return(nil).Once()

		err := retrier.Run(mock.count, "test")

		assert.Equal(t, 1, len(mock.Calls))
		assert.Nil(t, err)
	})

	t.Run("should exceeds the retry limiy when return an unexpected error 5 times", func(t *testing.T) {
		retrier := newRetrier(getExceededRateLimiter())
		mock := newMockCounter()
		mock.On("count").Return(errors.New(unexpectedErrMsg)).Times(5)

		err := retrier.Run(mock.count, "test")

		assert.Equal(t, 5, len(mock.Calls))
		assert.NotNil(t, err)
	})
}

type mockCounter struct {
	mock.Mock
}

func newMockCounter() *mockCounter {
	return &mockCounter{}
}

func (m *mockCounter) count() error {
	returnArg := m.MethodCalled("count")[0]
	if returnArg == nil {
		return nil
	}
	return returnArg.(error)
}

func getExceededRateLimiter() RateLimiter {
	mockTime := time.Now().Add(time.Second * -30)
	return newMockRateLimiter(mockGreaterLimit, mockTime)
}
