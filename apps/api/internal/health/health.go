// Package health defines the application readiness policy, independent of infrastructure.
package health

import (
	"context"
	"errors"
	"time"
)

type Checker interface{ Check(context.Context) error }

// Failure retains only bounded diagnostic metadata, never the underlying error
// (which can contain credentials, URLs, bucket names or database details).
type Failure struct {
	index  int
	reason string
}

func (f *Failure) Error() string        { return "dependencies unavailable" }
func (f *Failure) DependencyIndex() int { return f.index }
func (f *Failure) Reason() string       { return f.reason }

type Service struct {
	timeout time.Duration
	checks  []Checker
}

func New(timeout time.Duration, checks ...Checker) *Service {
	return &Service{timeout: timeout, checks: append([]Checker(nil), checks...)}
}
func (s *Service) Ready(ctx context.Context) error {
	ctx, cancel := context.WithTimeout(ctx, s.timeout)
	defer cancel()
	for index, check := range s.checks {
		if err := check.Check(ctx); err != nil {
			reason := "unavailable"
			if errors.Is(err, context.DeadlineExceeded) {
				reason = "timeout"
			} else if errors.Is(err, context.Canceled) {
				reason = "canceled"
			}
			return &Failure{index: index, reason: reason}
		}
	}
	return nil
}
