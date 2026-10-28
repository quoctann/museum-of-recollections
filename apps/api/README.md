# API

Go/Gin first slice: `GET /healthz` reports process health; `GET /readyz`
checks PostgreSQL connectivity and S3 bucket access in a shared 2-second
deadline. Healthy responses are 200; readiness failures are generic 503s.
No dependency credentials, addresses, or raw errors are returned or logged.

Article persistence is now available internally: goose migrations, sqlc queries,
version-checked draft saves and transactional publication snapshots. **Article
HTTP routes are not registered**, and the website still serves demo fixtures.
See [persistence operations](../../docs/persistence.md) and the
[proposed HTTP contract](../../docs/api-contract.md). Admin/auth/media tables do
not imply that login, sessions or upload are implemented.

## Run

Provide configuration through your environment/secret manager, then run
`go run ./cmd/api` from this directory. No environment files are included.

| Environment variable | Requirement/default |
| --- | --- |
| `API_PORT` | `8080`, range 1–65535 |
| `DATABASE_URL` | Required PostgreSQL URL (`postgres://` or `postgresql://`) |
| `S3_ENDPOINT` | Required `host:port`, no URL scheme or path |
| `S3_ACCESS_KEY` / `S3_SECRET_KEY` | Required credentials |
| `S3_BUCKET` | `recollections` |
| `S3_USE_SSL` | `false`; enable for TLS-backed storage |
| `SHUTDOWN_TIMEOUT` | `10s`; positive duration up to `1m` |

**Provision the configured bucket before expecting readiness.** This service
does not initialize or create buckets. Its storage identity must be allowed
to check that bucket (S3 HeadBucket/ListBucket permission). Readiness fails
for missing buckets, denied access, network errors, or database failures.
Liveness stays independent of dependencies. Database connections are opened
lazily, allowing a degraded process to start and report unready.

The local one-command stack provisions storage with the separate one-shot
`storage-init` executable, not through API startup. It validates only the S3
variables above (no database or API port required), then checks/creates the
bucket with a shared 30-second deadline. Existing buckets and concurrent
`BucketAlreadyOwnedByYou` responses succeed; access errors and buckets owned
by another identity fail. It never sets a public policy or modifies an existing
bucket's policy. A newly created bucket uses the storage service's private
default. Run manually with `go run ./cmd/storage-init`, or override the image's
entrypoint to `/usr/local/bin/storage-init`; Compose can gate API startup on
its `service_completed_successfully`. The provisioning identity needs bucket
creation permission in addition to readiness access.

Liveness, readiness, and other routes have independent bounded, process-local
token buckets (20 requests/second, burst 40 each); excess traffic receives 429
and Retry-After. Unknown-route traffic cannot consume either probe budget.
Direct flooding of a probe can still exhaust that probe's budget; restrict probe
access at ingress/network level and use ingress limits for per-client fairness.
No client-IP map or background cleanup goroutine is needed.
Readiness diagnostics contain only a dependency index and allowlisted reason;
HTTP responses remain generic and raw SDK/database errors are never logged.
Readiness and request timeouts use cooperative contexts, not detached handler
goroutines. HTTP header/read/write/idle timeouts are also configured. SIGINT
and SIGTERM drain active requests; timed-out drains force-close connections.

Build container from this directory: `docker build -t recollections-api .`.
The multi-stage image runs as UID/GID 10001. Deploy behind TLS; use database
TLS and `S3_USE_SSL=true` outside a trusted local network. No auth or user-data
routes are implemented in this slice.

## Tests and verification

Current persistence verification: `make test-integration` runs the pinned goose
up → up → down → up cycle and the Go race suite on a disposable PostgreSQL server.
The handwritten `internal/articles` package currently reaches **89.9% statement
coverage** with unit + real-database tests; this is not overall API coverage.
`make sqlc-check` passes. Regular unit tests skip database integration unless
`TEST_DATABASE_URL` is explicitly supplied; use the disposable runner by default.

### Historical foundation verification (before persistence)

Tests were authored before implementation. Initial `go test -mod=mod ./...`
failed with undefined `Load`, `Checker`, `New`, `Database`, and `Storage`.
Separate startup/lifecycle tests then failed with undefined `run` and `Serve`.
Implementations made these tests pass.

```sh
gofmt -w cmd internal
go mod tidy
go test -race -coverprofile=/tmp/opencode/museum-api-coverage.out ./...
go vet ./...
go tool cover -func=/tmp/opencode/museum-api-coverage.out
```

Race tests and vet pass with **85.9% overall statement coverage**; config,
storage initializer, dependency adapters, health, and HTTP handlers each reach
100%. Command packages remain lower because OS-exit/signal wrappers are not
invoked directly by unit tests. Storage bootstrap tests cover existing/new
buckets, ownership races, denied access, bounded contexts, and sanitized errors.
Regression tests cover cross-route rate-budget isolation and log redaction.

Go and the Docker builder require **1.27.3**, with Alpine 3.23.4 runtime and
both executables running as UID/GID 10001. Dependency scanning reports zero
reachable/imported-package vulnerabilities. An advisory remains at module level
for unused `x/crypto/openpgp` with no fixed version; this is not a claim that
every transitive module or the separate local MinIO image is vulnerability-free.

Coverage includes config validation/defaults, application health policy,
real MinIO client requests to an HTTP test server (success/missing/denied/
server errors/cancellation), pgx handshake timeout against a local socket,
handler responses/redaction/rate limits/concurrency, startup, and HTTP server
draining/force-close. Automated tests use isolated fakes and local HTTP/TCP
servers. Manual local integration verified PostgreSQL connectivity, two
successful initializer runs against MinIO (idempotency), and anonymous bucket
access returning 403. These are historical foundation results, before the article
persistence slice; upload remains unimplemented.
Container smoke checks also covered nonroot execution, health 200, generic
readiness 503 with unavailable dependencies, and graceful SIGTERM exit code 0.
