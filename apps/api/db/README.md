# Article persistence (phase A)

From `apps/api`, generate with the pinned tool (do not hand-edit generated Go):

```sh
go run github.com/sqlc-dev/sqlc/cmd/sqlc@v1.29.0 generate
```

Migration tooling is pinned to **goose v3.24.1**. Migrations run separately from
the API process. Supply a dedicated migration role/URL through the environment;
never commit credentials. Take a backup before destructive rollback.

```sh
python3 ../../scripts/api_tools.py up
python3 ../../scripts/api_tools.py down
```

The first `down` drops all five business tables and their data. The migration is
transactional; it creates no accounts. Media metadata is reserved for a later
slice: it grants no bucket access and does not enable image content.
The wrapper invokes pinned goose with `GOOSE_DBSTRING` in its environment, not
as a command-line argument. `DATABASE_URL` must already be set by your trusted
environment/secret manager. See [operations](../../../docs/persistence.md).

## Tests

```sh
go test ./internal/articles ./db/...
TEST_DATABASE_URL='postgres://…/postgres?sslmode=…' \
  go test -race -coverpkg=./internal/articles,./db/generated \
  -coverprofile=coverage.out ./internal/articles
```

Database tests **skip** unless `TEST_DATABASE_URL` is explicitly supplied. They
require a PostgreSQL URL and a role allowed to create/drop databases. Each test
creates a unique database, runs migration up/down/up, and force-drops **only that
database** during cleanup. They never discover `.env` or use application data.
Tests execute the embedded migration SQL after removing goose Up/Down markers;
they are not an embedded production migration runner. Goose CLI compatibility
is also verified separately with the pinned tool.

## Contract

`internal/articles.Repository` exposes create, get draft, save, publish,
unpublish, public detail and public list. `NewPostgres(*sqlx.DB)` implements it
with generated `database/sql` queries. The driver must already be registered
(the application currently uses pgx stdlib). No additional runtime dependencies
are required. Callers own connection lifecycle and authorization.

- Article IDs/revisions are positive int64 values. Creation starts at revision 1.
- Save, publish and unpublish require `expectedVersion`, lock the aggregate row,
  and increment revision once on success, even for repeated publish/unpublish.
  Any error rolls back the entire operation.
- Publish accepts no content; it copies the exact saved JSONB/title/subtitle in
  the same transaction as visibility/revision. JSONB preserves JSON values, not
  whitespace or object key ordering. Draft edits never modify that snapshot.
  The persisted draft is row-locked and revalidated before snapshotting, so SQL
  imports cannot bypass the content/size allowlist at publication time.
  Revalidation compacts JSONB-returned whitespace before checking size; Create/Save
  still enforce the original request-content byte limit. A valid 1 MiB input
  remains publishable after PostgreSQL's formatting expansion.
- Slugs match `[a-z0-9]+(-[a-z0-9]+)*`, maximum 120 ASCII characters. They are
  globally unique, including unpublished articles. They may change before first
  publication only. First-publication timestamp and slug remain immutable even
  after unpublish (also enforced by a database trigger). No redirects are needed.
- Unpublish hides but retains the latest snapshot. Republish refreshes the
  snapshot/publication timestamp, preserving first-publication timestamp.
- Public list returns document-free summaries, ordered by first-publication
  timestamp DESC, then ID DESC. `PublicList(ctx, limit, after)` accepts zero for
  the default limit 20, maximum 50, and an optional typed `PublicCursor` containing
  a positive bigint ID and a nonzero microsecond-precision timestamp (year 1–9999).
  `PublicPage` contains `Items` and `NextCursor` (nil at the end). Pagination uses
  a strict timestamp/ID tuple keyset, never OFFSET; republishing cannot reorder
  articles. Future HTTP transport must encode/decode opaque cursors safely.
  This is not snapshot isolation across requests; unpublishing may remove items.
- Sentinel errors: `ErrInvalid`, `ErrNotFound`, `ErrConflict`, `ErrSlugTaken`.
  Infrastructure errors are returned to the application boundary: never serialize
  them to clients. No HTTP handlers/auth wiring are included here.

Content is an object `{ "schemaVersion": 1, "doc": <Tiptap doc> }`, at most
1 MiB, 10,000 nodes and 32 node levels. Unknown/duplicate keys and isolated UTF-16
surrogate escapes are rejected; valid surrogate pairs are accepted.
Supported nodes: doc, paragraph, heading (levels 2/3), blockquote, bulletList,
orderedList (optional start 1–10000), listItem, text and hardBreak. Supported
marks: bold, italic, link (href only). Links permit HTTP(S), bare mailto addresses,
root-relative paths and nonempty fragments; network paths, credentials, control
characters and unsafe schemes are rejected. HTML, embeds, images, arbitrary
attributes and unsupported marks are rejected. An empty doc is permitted.
Titles are nonblank, at most 200 Unicode codepoints; subtitles at most 500.
No renderer is provided: future rendering must independently enforce this
allowlist and escape text/attributes rather than using raw HTML.
