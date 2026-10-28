package server

import (
	"context"
	"errors"
	"net"
	"net/http"
	"time"
)

// Serve waits for cancellation or a listener failure and always joins its serving
// goroutine. A timed-out drain force-closes active requests.
func Serve(ctx context.Context, srv *http.Server, ln net.Listener, timeout time.Duration) error {
	done := make(chan error, 1)
	go func() { done <- srv.Serve(ln) }()
	select {
	case err := <-done:
		if errors.Is(err, http.ErrServerClosed) {
			return nil
		}
		return err
	case <-ctx.Done():
		drain, cancel := context.WithTimeout(context.Background(), timeout)
		defer cancel()
		err := srv.Shutdown(drain)
		if err != nil {
			_ = srv.Close()
		}
		serveErr := <-done
		if err != nil {
			return err
		}
		if !errors.Is(serveErr, http.ErrServerClosed) {
			return serveErr
		}
		return nil
	}
}
