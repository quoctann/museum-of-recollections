package health

import (
	"context"
	"errors"
	"testing"
	"time"
)

type checkFunc func(context.Context) error

func (f checkFunc) Check(c context.Context) error { return f(c) }
func TestReadiness(t *testing.T) {
	for _, fail := range []int{-1, 0, 1} {
		calls := 0
		checks := []Checker{}
		for i := 0; i < 2; i++ {
			checks = append(checks, checkFunc(func(ctx context.Context) error {
				calls++
				if _, ok := ctx.Deadline(); !ok {
					t.Error("missing deadline")
				}
				if i == fail {
					return errors.New("private detail")
				}
				return nil
			}))
		}
		err := New(time.Second, checks...).Ready(context.Background())
		if (err != nil) != (fail >= 0) {
			t.Fatalf("unexpected result %v", err)
		}
		if fail < 0 && calls != 2 {
			t.Fatal("not all checked")
		}
	}
}
func TestReadinessTimeout(t *testing.T) {
	s := New(10*time.Millisecond, checkFunc(func(ctx context.Context) error { <-ctx.Done(); return ctx.Err() }))
	start := time.Now()
	if s.Ready(context.Background()) == nil {
		t.Fatal("expected timeout")
	}
	if time.Since(start) > time.Second {
		t.Fatal("unbounded")
	}
}

func TestFailureRetainsOnlySafeDiagnostics(t *testing.T) {
	for _, tc := range []struct {
		err    error
		reason string
	}{
		{errors.New("private password and endpoint"), "unavailable"},
		{context.DeadlineExceeded, "timeout"},
		{context.Canceled, "canceled"},
	} {
		s := New(time.Second,
			checkFunc(func(context.Context) error { return nil }),
			checkFunc(func(context.Context) error { return tc.err }))
		err := s.Ready(context.Background())
		var failure *Failure
		if !errors.As(err, &failure) {
			t.Fatalf("expected safe failure: %v", err)
		}
		if failure.DependencyIndex() != 1 || failure.Reason() != tc.reason || err.Error() != "dependencies unavailable" {
			t.Fatalf("unexpected diagnostics: %#v", failure)
		}
		if errors.Unwrap(err) != nil {
			t.Fatal("underlying error retained")
		}
	}
}
