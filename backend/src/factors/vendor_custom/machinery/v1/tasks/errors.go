package tasks

import (
	"fmt"
	"time"
)

// ErrRetryTaskLater ...
type ErrRetryTaskLater struct {
	name, msg string
	retryIn   time.Duration
}

type ErrRetryTaskExp ErrRetryTaskLater

// RetryIn returns time.Duration from now when task should be retried
func (e ErrRetryTaskLater) RetryIn() time.Duration {
	return e.retryIn
}

// Error implements the error interface
func (e ErrRetryTaskLater) Error() string {
	return fmt.Sprintf("Task error: %s Will retry in: %s", e.msg, e.retryIn)
}

// NewErrRetryTaskLater returns new ErrRetryTaskLater instance
func NewErrRetryTaskLater(msg string, retryIn time.Duration) ErrRetryTaskLater {
	return ErrRetryTaskLater{msg: msg, retryIn: retryIn}
}

// RetryIn returns time.Duration from now when task should be retried
func (e ErrRetryTaskExp) RetryIn() time.Duration {
	return e.retryIn
}

// Error implements the error interface
func (e ErrRetryTaskExp) Error() string {
	return fmt.Sprintf("Task error: %s Will retry in: %s", e.msg, e.retryIn)
}

// NewErrRetryTaskExp returns new ErrRetryTaskExp instance
func NewErrRetryTaskExp(msg string) ErrRetryTaskExp {
	return ErrRetryTaskExp{msg: msg}
}

// Retriable is interface that retriable errors should implement
type Retriable interface {
	RetryIn() time.Duration
}
