package main

import (
	"context"
	"fmt"
	"go.uber.org/zap"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestInvalidConfig(t *testing.T) {
	if run(context.Background(), func(string) string { return "" }, zap.NewNop()) == nil {
		t.Fatal("expected config error")
	}
}
func TestRun(t *testing.T) {
	for _, tc := range []struct {
		name      string
		head, put int
		code      string
		fail      bool
	}{
		{"existing", 200, 0, "", false}, {"create", 404, 200, "", false}, {"race", 404, 409, "BucketAlreadyOwnedByYou", false}, {"other owner", 404, 409, "BucketAlreadyExists", true}, {"denied", 403, 0, "AccessDenied", true},
	} {
		t.Run(tc.name, func(t *testing.T) {
			puts := 0
			srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				if r.URL.Path != "/recollections/" {
					t.Errorf("unexpected path %s", r.URL.Path)
				}
				if r.Method == "HEAD" {
					w.WriteHeader(tc.head)
					return
				}
				if r.Method != "PUT" || r.URL.RawQuery != "" {
					t.Errorf("unexpected mutation %s %s", r.Method, r.URL)
				}
				puts++
				w.Header().Set("Content-Type", "application/xml")
				w.WriteHeader(tc.put)
				if tc.code != "" {
					fmt.Fprintf(w, "<Error><Code>%s</Code><Message>private detail</Message></Error>", tc.code)
				}
			}))
			defer srv.Close()
			env := map[string]string{"S3_ENDPOINT": strings.TrimPrefix(srv.URL, "http://"), "S3_ACCESS_KEY": "test", "S3_SECRET_KEY": "test"}
			err := run(context.Background(), func(k string) string { return env[k] }, zap.NewNop())
			if (err != nil) != tc.fail {
				t.Fatalf("error %v", err)
			}
			if err != nil && strings.Contains(err.Error(), "private") {
				t.Fatal("unsafe error")
			}
			if (puts > 0) != (tc.put > 0) {
				t.Fatalf("puts %d", puts)
			}
		})
	}
}
