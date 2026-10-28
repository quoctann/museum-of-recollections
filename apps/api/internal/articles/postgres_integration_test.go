package articles

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/url"
	"os"
	"reflect"
	"strings"
	"sync"
	"testing"
	"time"

	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/jmoiron/sqlx"
	migrations "museum-of-recollections/apps/api/db"
)

func testDatabase(t *testing.T) *sqlx.DB {
	t.Helper()
	raw := os.Getenv("TEST_DATABASE_URL")
	if raw == "" {
		t.Skip("TEST_DATABASE_URL required (role must CREATE DATABASE)")
	}
	u, err := url.Parse(raw)
	if err != nil || (u.Scheme != "postgres" && u.Scheme != "postgresql") || u.Host == "" {
		t.Fatal("invalid TEST_DATABASE_URL")
	}
	admin, err := sqlx.Connect("pgx", raw)
	if err != nil {
		t.Fatal("connect test maintenance database failed")
	}
	name := fmt.Sprintf("articles_test_%d", time.Now().UnixNano())
	if _, err = admin.Exec(`CREATE DATABASE "` + name + `"`); err != nil {
		admin.Close()
		t.Fatal(err)
	}
	var db *sqlx.DB
	t.Cleanup(func() {
		if db != nil {
			db.Close()
		}
		if _, err := admin.Exec(`DROP DATABASE "` + name + `" WITH (FORCE)`); err != nil {
			t.Error("failed to remove isolated test database")
		}
		admin.Close()
	})
	db, err = sqlx.Connect("pgx", isolatedDatabaseURL(u, name))
	if err != nil {
		t.Fatal(err)
	}
	data, err := migrations.Migrations.ReadFile("migrations/00001_articles.sql")
	if err != nil {
		t.Fatal(err)
	}
	parts := strings.Split(string(data), "-- +goose Down")
	up := strings.ReplaceAll(parts[0], "-- +goose Up", "")
	down := parts[1]
	for _, sql := range []string{up, down, up} {
		if _, err = db.Exec(sql); err != nil {
			t.Fatal(err)
		}
	}
	return db
}

func TestPostgresPublishContentSizeBoundary(t *testing.T) {
	db := testDatabase(t)
	p := NewPostgres(db)
	ctx := context.Background()
	prefix := `{"schemaVersion":1,"doc":{"type":"doc","content":[{"type":"paragraph","content":[{"type":"text","text":"`
	suffix := `"}]}]}}`
	in := testInput("size-boundary", "Boundary")
	in.Content = json.RawMessage(prefix + strings.Repeat("x", MaxContentBytes-len(prefix)-len(suffix)) + suffix)
	if len(in.Content) != MaxContentBytes {
		t.Fatal("incorrect boundary fixture")
	}
	d, err := p.Create(ctx, in)
	if err != nil {
		t.Fatal(err)
	}
	if len(d.Content) <= MaxContentBytes {
		t.Fatal("fixture did not reproduce JSONB whitespace expansion")
	}
	d, err = p.Save(ctx, d.ID, d.Version, in)
	if err != nil {
		t.Fatal(err)
	}
	d, err = p.Publish(ctx, d.ID, d.Version)
	if err != nil {
		t.Fatal("valid compact draft cannot publish after JSONB expansion", err)
	}
	before, err := p.PublicDetail(ctx, in.Slug)
	if err != nil {
		t.Fatal(err)
	}
	oversized := in
	oversized.Content = json.RawMessage(prefix + strings.Repeat("x", MaxContentBytes-len(prefix)-len(suffix)+1) + suffix)
	if _, err = p.Save(ctx, d.ID, d.Version, oversized); !errors.Is(err, ErrInvalid) {
		t.Fatal("raw over-limit save accepted", err)
	}
	oversized.Slug = "oversized-create"
	if _, err = p.Create(ctx, oversized); !errors.Is(err, ErrInvalid) {
		t.Fatal("raw over-limit create accepted", err)
	}
	// An SQL import bypasses raw input checks, but a truly over-limit compact
	// document must still be rejected at publish, without changing the snapshot.
	if _, err = db.Exec(`UPDATE article_drafts SET content=$1 WHERE article_id=$2`, oversized.Content, d.ID); err != nil {
		t.Fatal(err)
	}
	if _, err = p.Publish(ctx, d.ID, d.Version); !errors.Is(err, ErrInvalid) {
		t.Fatal("over-limit compact import published", err)
	}
	got, err := p.GetDraft(ctx, d.ID)
	if err != nil || got.Version != d.Version || !got.Published {
		t.Fatal("rejected publish changed aggregate", err)
	}
	after, err := p.PublicDetail(ctx, in.Slug)
	if err != nil || !bytes.Equal(before.Content, after.Content) || !before.PublishedAt.Equal(after.PublishedAt) {
		t.Fatal("rejected publish changed snapshot", err)
	}
}

func TestIsolatedDatabaseURL(t *testing.T) {
	u, err := url.Parse("postgres://test@localhost/maintenance?dbname=unsafe&database=unsafe&sslmode=disable")
	if err != nil {
		t.Fatal(err)
	}
	got, err := url.Parse(isolatedDatabaseURL(u, "articles_test_isolated"))
	if err != nil {
		t.Fatal(err)
	}
	if got.Path != "/articles_test_isolated" || got.Query().Has("dbname") || got.Query().Has("database") || got.Query().Get("sslmode") != "disable" || u.Path != "/maintenance" {
		t.Fatal("test database URL was not isolated")
	}
}

// pgx query parameters can override the path database. Never let a maintenance
// URL's dbname/database parameter redirect destructive migrations to that DB.
func isolatedDatabaseURL(maintenance *url.URL, name string) string {
	u := *maintenance
	u.Path = "/" + name
	u.RawPath = ""
	query := u.Query()
	query.Del("dbname")
	query.Del("database")
	u.RawQuery = query.Encode()
	return u.String()
}

func testInput(slug, title string) Input {
	return Input{Title: title, Slug: slug, Content: json.RawMessage(`{"schemaVersion":1,"doc":{"type":"doc","content":[{"type":"paragraph","content":[{"type":"text","text":"Memory"}]}]}}`)}
}

func TestPostgresLifecycle(t *testing.T) {
	db := testDatabase(t)
	p := NewPostgres(db)
	ctx := context.Background()
	d, err := p.Create(ctx, testInput("memory", "First"))
	if err != nil {
		t.Fatal(err)
	}
	if d.Version != 1 || d.Published || d.FirstPublishedAt != nil {
		t.Fatalf("bad initial draft: %+v", d)
	}
	if _, err = p.PublicDetail(ctx, "memory"); !errors.Is(err, ErrNotFound) {
		t.Fatal(err)
	}
	pubDraft, err := p.Publish(ctx, d.ID, d.Version)
	if err != nil {
		t.Fatal(err)
	}
	if !pubDraft.Published || pubDraft.Version != 2 || pubDraft.FirstPublishedAt == nil {
		t.Fatal(pubDraft)
	}
	pub, err := p.PublicDetail(ctx, "memory")
	if err != nil {
		t.Fatal(err)
	}
	edited := testInput("memory", "Second")
	edited.Content = json.RawMessage(`{"schemaVersion":1,"doc":{"type":"doc","content":[{"type":"paragraph","content":[{"type":"text","text":"Edited snapshot 🌿"}]}]}}`)
	saved, err := p.Save(ctx, d.ID, 2, edited)
	if err != nil {
		t.Fatal(err)
	}
	old, err := p.PublicDetail(ctx, "memory")
	if err != nil || old.Title != "First" || string(old.Content) != string(pub.Content) {
		t.Fatal("draft changed publication", err)
	}
	for _, mutate := range []func() error{
		func() error { _, e := p.Save(ctx, d.ID, 2, edited); return e }, func() error { _, e := p.Publish(ctx, d.ID, 2); return e }, func() error { _, e := p.Unpublish(ctx, d.ID, 2); return e },
	} {
		if !errors.Is(mutate(), ErrConflict) {
			t.Fatal("stale mutation accepted")
		}
	}
	if _, err = p.Save(ctx, d.ID, saved.Version, testInput("changed", "Second")); !errors.Is(err, ErrInvalid) {
		t.Fatal("slug changed after publication", err)
	}
	republished, err := p.Publish(ctx, d.ID, saved.Version)
	if err != nil {
		t.Fatal(err)
	}
	latest, err := p.PublicDetail(ctx, "memory")
	if err != nil || latest.Title != "Second" || !latest.FirstPublishedAt.Equal(pub.FirstPublishedAt) {
		t.Fatal("bad republish", err)
	}
	var savedJSON, publicJSON any
	if json.Unmarshal(edited.Content, &savedJSON) != nil || json.Unmarshal(latest.Content, &publicJSON) != nil || !reflect.DeepEqual(savedJSON, publicJSON) {
		t.Fatal("publication did not snapshot exact saved content")
	}
	list, err := p.PublicList(ctx, 10, nil)
	if err != nil || len(list.Items) != 1 || list.Items[0].Title != "Second" {
		t.Fatal(list, err)
	}
	unpublished, err := p.Unpublish(ctx, d.ID, republished.Version)
	if err != nil || unpublished.Published {
		t.Fatal(err)
	}
	if _, err = p.PublicDetail(ctx, "memory"); !errors.Is(err, ErrNotFound) {
		t.Fatal(err)
	}
	list, err = p.PublicList(ctx, 10, nil)
	if err != nil || len(list.Items) != 0 {
		t.Fatal(list, err)
	}
	if _, err = p.Save(ctx, d.ID, unpublished.Version, testInput("changed", "Second")); !errors.Is(err, ErrInvalid) {
		t.Fatal(err)
	}
	if _, err = p.Create(ctx, testInput("memory", "Other")); !errors.Is(err, ErrSlugTaken) {
		t.Fatal(err)
	}
	if _, err = p.GetDraft(ctx, 999999); !errors.Is(err, ErrNotFound) {
		t.Fatal(err)
	}
	got, err := p.GetDraft(ctx, d.ID)
	if err != nil || got.Version != unpublished.Version {
		t.Fatal(err)
	}
}

func TestPostgresRacesAndRollback(t *testing.T) {
	db := testDatabase(t)
	p := NewPostgres(db)
	ctx := context.Background()
	d, err := p.Create(ctx, testInput("race", "Race"))
	if err != nil {
		t.Fatal(err)
	}
	var wg sync.WaitGroup
	out := make(chan error, 2)
	for range 2 {
		wg.Add(1)
		go func() { defer wg.Done(); _, e := p.Save(ctx, d.ID, 1, testInput("race", "Winner")); out <- e }()
	}
	wg.Wait()
	close(out)
	successes, conflicts := 0, 0
	for e := range out {
		if e == nil {
			successes++
		} else if errors.Is(e, ErrConflict) {
			conflicts++
		} else {
			t.Fatal(e)
		}
	}
	if successes != 1 || conflicts != 1 {
		t.Fatal(successes, conflicts)
	}
	_, err = db.Exec(`CREATE FUNCTION reject_publication() RETURNS trigger LANGUAGE plpgsql AS $$ BEGIN RAISE EXCEPTION 'forced failure'; END $$; CREATE TRIGGER reject_publication BEFORE INSERT OR UPDATE ON article_publications FOR EACH ROW EXECUTE FUNCTION reject_publication()`)
	if err != nil {
		t.Fatal(err)
	}
	if _, err = p.Publish(ctx, d.ID, 2); err == nil {
		t.Fatal("forced failure ignored")
	}
	got, err := p.GetDraft(ctx, d.ID)
	if err != nil || got.Version != 2 || got.Published || got.FirstPublishedAt != nil {
		t.Fatalf("partial publish: %+v %v", got, err)
	}
	if _, err = p.PublicDetail(ctx, "race"); !errors.Is(err, ErrNotFound) {
		t.Fatal(err)
	}
	// Fail the aggregate UPDATE after the snapshot INSERT succeeded, proving
	// rollback covers both halves of publish, not merely a failed first write.
	_, err = db.Exec(`DROP TRIGGER reject_publication ON article_publications;
CREATE FUNCTION reject_article_publish() RETURNS trigger LANGUAGE plpgsql AS $$ BEGIN IF NEW.published THEN RAISE EXCEPTION 'forced aggregate failure'; END IF; RETURN NEW; END $$;
CREATE TRIGGER reject_article_publish BEFORE UPDATE ON articles FOR EACH ROW EXECUTE FUNCTION reject_article_publish()`)
	if err != nil {
		t.Fatal(err)
	}
	if _, err = p.Publish(ctx, d.ID, 2); err == nil {
		t.Fatal("aggregate failure ignored")
	}
	var count int
	if err = db.Get(&count, `SELECT count(*) FROM article_publications WHERE article_id=$1`, d.ID); err != nil || count != 0 {
		t.Fatal("snapshot survived rollback", err)
	}
}

func TestPostgresConstraintsAndDraftSlug(t *testing.T) {
	db := testDatabase(t)
	p := NewPostgres(db)
	ctx := context.Background()
	d, err := p.Create(ctx, testInput("before", "Before"))
	if err != nil {
		t.Fatal(err)
	}
	changed, err := p.Save(ctx, d.ID, 1, testInput("after", "After"))
	if err != nil || changed.Slug != "after" || changed.Version != 2 {
		t.Fatal(changed, err)
	}
	other, err := p.Create(ctx, testInput("taken", "Other"))
	if err != nil {
		t.Fatal(err)
	}
	if _, err = p.Save(ctx, d.ID, 2, testInput(other.Slug, "Must rollback")); !errors.Is(err, ErrSlugTaken) {
		t.Fatal(err)
	}
	got, err := p.GetDraft(ctx, d.ID)
	if err != nil || got.Title != "After" || got.Version != 2 {
		t.Fatal("partial save on duplicate slug", got, err)
	}
	for _, content := range []string{`{"schemaVersion":1}`, `{"schemaVersion":1,"doc":null}`, `{"doc":{"type":"doc"}}`} {
		if _, err = db.Exec(`UPDATE article_drafts SET content=$1 WHERE article_id=$2`, content, d.ID); err == nil {
			t.Fatalf("schema accepted %s", content)
		}
	}
	pub, err := p.Publish(ctx, d.ID, 2)
	if err != nil {
		t.Fatal(err)
	}
	if _, err = db.Exec(`UPDATE articles SET slug='illegal' WHERE id=$1`, d.ID); err == nil {
		t.Fatal("database allowed published slug mutation")
	}
	if _, err = db.Exec(`UPDATE articles SET first_published_at=NULL WHERE id=$1`, d.ID); err == nil {
		t.Fatal("database allowed identity reset")
	}
	if _, err = p.Unpublish(ctx, d.ID, pub.Version); err != nil {
		t.Fatal(err)
	}
	if _, err = db.Exec(`UPDATE articles SET slug='illegal' WHERE id=$1`, d.ID); err == nil {
		t.Fatal("database allowed unpublished slug mutation")
	}
}

func TestPostgresInvalidAndMissing(t *testing.T) {
	db := testDatabase(t)
	p := NewPostgres(db)
	ctx := context.Background()
	for _, call := range []func() error{
		func() error { _, e := p.Create(ctx, Input{}); return e },
		func() error { _, e := p.GetDraft(ctx, 0); return e },
		func() error { _, e := p.Save(ctx, 1, 1, Input{}); return e },
		func() error { _, e := p.Publish(ctx, 1, 0); return e },
		func() error { _, e := p.Unpublish(ctx, 0, 1); return e },
		func() error { _, e := p.PublicDetail(ctx, "Bad"); return e },
		func() error { _, e := p.PublicList(ctx, -1, nil); return e },
		func() error { _, e := p.PublicList(ctx, 51, nil); return e },
		func() error {
			_, e := p.PublicList(ctx, 1, &PublicCursor{ID: -1, FirstPublishedAt: time.Now()})
			return e
		},
		func() error { _, e := p.PublicList(ctx, 1, &PublicCursor{ID: 1}); return e },
	} {
		if !errors.Is(call(), ErrInvalid) {
			t.Fatal("invalid input accepted")
		}
	}
	for _, call := range []func() error{
		func() error { _, e := p.Save(ctx, 999, 1, testInput("missing", "Missing")); return e },
		func() error { _, e := p.Publish(ctx, 999, 1); return e },
		func() error { _, e := p.Unpublish(ctx, 999, 1); return e },
	} {
		if !errors.Is(call(), ErrNotFound) {
			t.Fatal("missing article accepted")
		}
	}
	canceled, cancel := context.WithCancel(ctx)
	cancel()
	if _, err := p.Create(canceled, testInput("cancel", "Cancel")); err == nil {
		t.Fatal("canceled create succeeded")
	}
	if _, err := p.Publish(canceled, 1, 1); err == nil {
		t.Fatal("canceled publish succeeded")
	}
	db.Close()
	if _, err := p.GetDraft(ctx, 1); err == nil {
		t.Fatal("closed database read succeeded")
	}
	if _, err := p.PublicDetail(ctx, "x"); err == nil {
		t.Fatal("closed public read succeeded")
	}
	if _, err := p.PublicList(ctx, 1, nil); err == nil {
		t.Fatal("closed list succeeded")
	}
}

func TestPostgresPublishSaveRaceAndPagination(t *testing.T) {
	db := testDatabase(t)
	p := NewPostgres(db)
	ctx := context.Background()
	d, err := p.Create(ctx, testInput("concurrent", "Original"))
	if err != nil {
		t.Fatal(err)
	}
	start := make(chan struct{})
	results := make(chan error, 2)
	go func() { <-start; _, e := p.Publish(ctx, d.ID, 1); results <- e }()
	go func() { <-start; _, e := p.Save(ctx, d.ID, 1, testInput("concurrent", "Edited")); results <- e }()
	close(start)
	first, second := <-results, <-results
	if !((first == nil && errors.Is(second, ErrConflict)) || (second == nil && errors.Is(first, ErrConflict))) {
		t.Fatal(first, second)
	}
	got, err := p.GetDraft(ctx, d.ID)
	if err != nil || got.Version != 2 {
		t.Fatal(got, err)
	}
	if got.Published {
		pub, e := p.PublicDetail(ctx, "concurrent")
		if e != nil || pub.Title != "Original" || got.Title != "Original" {
			t.Fatal("publication mixed concurrent versions", e)
		}
	} else if got.Title != "Edited" {
		t.Fatal("winning save missing")
	}
	// Normalize visibility, then create a second publication to verify ordering
	// and bounded pagination independent of which concurrent operation won.
	if _, err = p.Publish(ctx, d.ID, 2); err != nil {
		t.Fatal(err)
	}
	newer, err := p.Create(ctx, testInput("newer", "Newer"))
	if err != nil {
		t.Fatal(err)
	}
	if _, err = p.Publish(ctx, newer.ID, 1); err != nil {
		t.Fatal(err)
	}
	page, err := p.PublicList(ctx, 1, nil)
	if err != nil || len(page.Items) != 1 || page.Items[0].ID != newer.ID || page.NextCursor == nil {
		t.Fatal(page, err)
	}
	page, err = p.PublicList(ctx, 1, page.NextCursor)
	if err != nil || len(page.Items) != 1 || page.Items[0].ID != d.ID || page.NextCursor != nil {
		t.Fatal(page, err)
	}
	page, err = p.PublicList(ctx, 1, &PublicCursor{ID: d.ID, FirstPublishedAt: page.Items[0].FirstPublishedAt})
	if err != nil || page.Items == nil || len(page.Items) != 0 {
		t.Fatal(page, err)
	}
}

func TestPostgresTiedPublicationOrder(t *testing.T) {
	db := testDatabase(t)
	p := NewPostgres(db)
	ctx := context.Background()
	a, err := p.Create(ctx, testInput("tie-a", "A"))
	if err != nil {
		t.Fatal(err)
	}
	b, err := p.Create(ctx, testInput("tie-b", "B"))
	if err != nil {
		t.Fatal(err)
	}
	// Both identities are still unpublished, so assign a shared first timestamp
	// as an import fixture without disabling the immutable-identity trigger.
	stamp := time.Date(2026, 1, 1, 0, 0, 0, 123456000, time.UTC)
	if _, err = db.Exec(`UPDATE articles SET first_published_at=$1 WHERE id IN ($2,$3)`, stamp, a.ID, b.ID); err != nil {
		t.Fatal(err)
	}
	a, err = p.Publish(ctx, a.ID, 1)
	if err != nil {
		t.Fatal(err)
	}
	if _, err = p.Publish(ctx, b.ID, 1); err != nil {
		t.Fatal(err)
	}
	page, err := p.PublicList(ctx, 1, nil)
	if err != nil || len(page.Items) != 1 || page.Items[0].ID != b.ID || page.NextCursor == nil {
		t.Fatal(page, err)
	}
	encoded, err := json.Marshal(page.Items[0])
	if err != nil || strings.Contains(string(encoded), `"content"`) {
		t.Fatal("summary exposed document", string(encoded), err)
	}
	cursor := page.NextCursor
	if _, err = p.Publish(ctx, a.ID, a.Version); err != nil {
		t.Fatal(err)
	}
	all, err := p.PublicList(ctx, 0, nil)
	if err != nil || len(all.Items) != 2 || all.Items[0].ID != b.ID || all.Items[1].ID != a.ID || !all.Items[1].FirstPublishedAt.Equal(stamp) {
		t.Fatal("republish reordered first-publication order", all, err)
	}
	next, err := p.PublicList(ctx, 1, cursor)
	if err != nil || len(next.Items) != 1 || next.Items[0].ID != a.ID || next.NextCursor != nil {
		t.Fatal("tie pagination failed", next, err)
	}
}

func TestPostgresPublishRevalidatesStoredDraft(t *testing.T) {
	db := testDatabase(t)
	p := NewPostgres(db)
	ctx := context.Background()
	d, err := p.Create(ctx, testInput("imported", "Safe"))
	if err != nil {
		t.Fatal(err)
	}
	d, err = p.Publish(ctx, d.ID, 1)
	if err != nil {
		t.Fatal(err)
	}
	before, err := p.PublicDetail(ctx, "imported")
	if err != nil {
		t.Fatal(err)
	}
	unsafe := `{"schemaVersion":1,"doc":{"type":"doc","content":[{"type":"paragraph","content":[{"type":"text","text":"unsafe","marks":[{"type":"link","attrs":{"href":"javascript:alert(1)"}}]}]}]}}`
	for _, content := range []string{unsafe, `{"schemaVersion":1,"doc":{"type":"rawHTML"}}`, `{"schemaVersion":1,"doc":{"type":"doc","content":[{"type":"paragraph","content":[{"type":"text","text":"` + strings.Repeat("x", MaxContentBytes) + `"}]}]}}`} {
		if _, err = db.Exec(`UPDATE article_drafts SET content=$1 WHERE article_id=$2`, content, d.ID); err != nil {
			t.Fatal(err)
		}
		if _, err = p.Publish(ctx, d.ID, d.Version); !errors.Is(err, ErrInvalid) {
			t.Fatal("unsafe persisted draft published", err)
		}
		got, e := p.GetDraft(ctx, d.ID)
		if e != nil || got.Version != d.Version || !got.Published {
			t.Fatal("rejection changed revision/state", got, e)
		}
		after, e := p.PublicDetail(ctx, "imported")
		if e != nil || !reflect.DeepEqual(before, after) {
			t.Fatal("rejection changed snapshot", e)
		}
	}
}
