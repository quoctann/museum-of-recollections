package dependencies

import (
	"context"
	"database/sql"
	"errors"
	"net"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
)

type pingFunc func(context.Context) error

func TestDatabasePingHonorsDeadline(t *testing.T) {
	ln, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}
	defer ln.Close()
	done := make(chan struct{})
	ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
	defer cancel()
	go func() {
		defer close(done)
		conn, err := ln.Accept()
		if err != nil {
			return
		}
		defer conn.Close()
		<-ctx.Done()
	}()
	db, err := sql.Open("pgx", "postgres://test:test@"+ln.Addr().String()+"/test?sslmode=disable")
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	if err := (Database{DB: db}).Check(ctx); err == nil {
		t.Fatal("expected handshake timeout")
	}
	ln.Close()
	select {
	case <-done:
	case <-time.After(time.Second):
		t.Fatal("listener leaked")
	}
}

func TestStorageHonorsDeadline(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) { <-r.Context().Done() }))
	defer srv.Close()
	client, err := minio.New(strings.TrimPrefix(srv.URL, "http://"), &minio.Options{Creds: credentials.NewStaticV4("test", "test", ""), Region: "us-east-1"})
	if err != nil {
		t.Fatal(err)
	}
	ctx, cancel := context.WithTimeout(context.Background(), 50*time.Millisecond)
	defer cancel()
	if err := (Storage{Client: client, Bucket: "recollections"}).Check(ctx); err == nil {
		t.Fatal("expected storage timeout")
	}
}

func (f pingFunc) PingContext(c context.Context) error { return f(c) }
func TestDatabase(t *testing.T) {
	for _, want := range []error{nil, errors.New("offline")} {
		d := Database{DB: pingFunc(func(context.Context) error { return want })}
		if got := d.Check(context.Background()); !errors.Is(got, want) {
			t.Fatalf("got %v", got)
		}
	}
}
func TestStorage(t *testing.T) {
	for _, status := range []int{200, 403, 404, 500} {
		t.Run(http.StatusText(status), func(t *testing.T) {
			srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				if r.Method != "HEAD" || r.URL.Path != "/recollections/" {
					t.Errorf("unexpected request %s %s", r.Method, r.URL.Path)
				}
				w.WriteHeader(status)
			}))
			defer srv.Close()
			client, err := minio.New(strings.TrimPrefix(srv.URL, "http://"), &minio.Options{Creds: credentials.NewStaticV4("test", "test", ""), Region: "us-east-1"})
			if err != nil {
				t.Fatal(err)
			}
			ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
			defer cancel()
			err = (Storage{Client: client, Bucket: "recollections"}).Check(ctx)
			if (err == nil) != (status == 200) {
				t.Fatalf("status %d: %v", status, err)
			}
		})
	}
}
