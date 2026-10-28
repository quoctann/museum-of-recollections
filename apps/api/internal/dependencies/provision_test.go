package dependencies

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
	"museum-of-recollections/apps/api/internal/storageinit"
)

func TestProvisioner(t *testing.T) {
	for _, code := range []string{"", "BucketAlreadyOwnedByYou", "BucketAlreadyExists"} {
		t.Run(code, func(t *testing.T) {
			srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				if r.URL.RawQuery != "" {
					t.Error("unexpected bucket policy operation")
				}
				if r.Method == "HEAD" {
					w.WriteHeader(200)
					return
				}
				if r.Method != "PUT" {
					t.Errorf("unexpected method %s", r.Method)
				}
				if code != "" {
					w.Header().Set("Content-Type", "application/xml")
					w.WriteHeader(409)
					fmt.Fprintf(w, "<Error><Code>%s</Code></Error>", code)
				}
			}))
			defer srv.Close()
			client, err := minio.New(strings.TrimPrefix(srv.URL, "http://"), &minio.Options{Creds: credentials.NewStaticV4("test", "test", ""), Region: "us-east-1"})
			if err != nil {
				t.Fatal(err)
			}
			p := Provisioner{Client: client}
			exists, err := p.Exists(context.Background(), "recollections")
			if err != nil || !exists {
				t.Fatalf("exists %v %v", exists, err)
			}
			err = p.Create(context.Background(), "recollections")
			if (err == nil) != (code == "") {
				t.Fatalf("unexpected error %v", err)
			}
			if errors.Is(err, storageinit.ErrAlreadyOwned) != (code == "BucketAlreadyOwnedByYou") {
				t.Fatalf("incorrect ownership translation: %v", err)
			}
		})
	}
}
