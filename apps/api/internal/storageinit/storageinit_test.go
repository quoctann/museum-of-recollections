package storageinit

import (
	"context"
	"errors"
	"strings"
	"testing"
	"time"
)

type fakeBucket struct {
	exists              bool
	checkErr, createErr error
	creates             int
}

func (f *fakeBucket) Exists(ctx context.Context, bucket string) (bool, error) {
	if _, ok := ctx.Deadline(); !ok {
		return false, errors.New("deadline missing")
	}
	return f.exists, f.checkErr
}
func (f *fakeBucket) Create(ctx context.Context, bucket string) error {
	f.creates++
	return f.createErr
}
func TestEnsure(t *testing.T) {
	for _, tc := range []struct {
		name                string
		exists              bool
		checkErr, createErr error
		creates             int
		fail                bool
	}{
		{name: "existing", exists: true}, {name: "new", creates: 1}, {name: "owned race", createErr: ErrAlreadyOwned, creates: 1}, {name: "check failure", checkErr: errors.New("private check"), fail: true}, {name: "creation failure", createErr: errors.New("private create"), creates: 1, fail: true},
	} {
		t.Run(tc.name, func(t *testing.T) {
			f := &fakeBucket{exists: tc.exists, checkErr: tc.checkErr, createErr: tc.createErr}
			err := Ensure(context.Background(), f, "recollections", time.Second)
			if (err != nil) != tc.fail || f.creates != tc.creates {
				t.Fatalf("error %v creates %d", err, f.creates)
			}
			if err != nil && strings.Contains(err.Error(), "private") {
				t.Fatal("unsafe error")
			}
		})
	}
}

type blockingBucket struct{}

func (blockingBucket) Exists(c context.Context, _ string) (bool, error) {
	<-c.Done()
	return false, c.Err()
}
func (blockingBucket) Create(context.Context, string) error { panic("unexpected create") }
func TestEnsureDeadline(t *testing.T) {
	start := time.Now()
	if Ensure(context.Background(), blockingBucket{}, "recollections", time.Millisecond) == nil {
		t.Fatal("expected failure")
	}
	if time.Since(start) > time.Second {
		t.Fatal("unbounded")
	}
}
