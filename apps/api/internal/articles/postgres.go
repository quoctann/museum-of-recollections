package articles

import (
	"bytes"
	"context"
	"database/sql"
	"encoding/json"
	"errors"

	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jmoiron/sqlx"
	q "museum-of-recollections/apps/api/db/generated"
)

// Postgres implements the use cases using sqlx-managed connections and sqlc's
// database/sql queries. All writes hold the aggregate row lock until commit.
type Postgres struct {
	db      *sqlx.DB
	queries *q.Queries
}

var _ Repository = (*Postgres)(nil)

func NewPostgres(db *sqlx.DB) *Postgres { return &Postgres{db: db, queries: q.New(db)} }

func (p *Postgres) Create(ctx context.Context, in Input) (Draft, error) {
	if err := ValidateInput(in); err != nil {
		return Draft{}, err
	}
	tx, err := p.db.BeginTxx(ctx, nil)
	if err != nil {
		return Draft{}, databaseError(err)
	}
	defer tx.Rollback()
	queries := q.New(tx)
	a, err := queries.CreateArticle(ctx, in.Slug)
	if err != nil {
		return Draft{}, databaseError(err)
	}
	// INSERT owns the newly-created aggregate row; no competing writer can see it.
	if err = queries.CreateDraft(ctx, q.CreateDraftParams{ArticleID: a.ID, Title: in.Title, Subtitle: in.Subtitle, Content: in.Content}); err != nil {
		return Draft{}, databaseError(err)
	}
	return commitDraft(ctx, tx, queries, a.ID)
}

func (p *Postgres) GetDraft(ctx context.Context, id int64) (Draft, error) {
	if id <= 0 {
		return Draft{}, ErrInvalid
	}
	d, err := p.queries.GetDraft(ctx, id)
	if err != nil {
		return Draft{}, databaseError(err)
	}
	return draftFromRow(d), nil
}

func (p *Postgres) Save(ctx context.Context, id, expectedVersion int64, in Input) (Draft, error) {
	if err := ValidateInput(in); err != nil {
		return Draft{}, err
	}
	return p.mutate(ctx, id, expectedVersion, func(queries *q.Queries, a q.Article) error {
		if a.FirstPublishedAt.Valid && a.Slug != in.Slug {
			return ErrInvalid
		}
		if err := queries.SaveDraft(ctx, q.SaveDraftParams{ArticleID: id, Title: in.Title, Subtitle: in.Subtitle, Content: in.Content}); err != nil {
			return err
		}
		return queries.ReviseArticle(ctx, q.ReviseArticleParams{ID: id, Slug: in.Slug})
	})
}

func (p *Postgres) Publish(ctx context.Context, id, expectedVersion int64) (Draft, error) {
	return p.mutate(ctx, id, expectedVersion, func(queries *q.Queries, _ q.Article) error {
		draft, err := queries.LockDraftForPublication(ctx, id)
		if err != nil {
			return err
		}
		// JSONB adds formatting whitespace on read. Apply the byte bound to its
		// compact representation, not that database-generated formatting. Raw
		// Create/Save requests still pass through the original strict byte limit.
		var compact bytes.Buffer
		if err = json.Compact(&compact, draft.Content); err != nil {
			return ErrInvalid
		}
		if err = ValidateInput(Input{Slug: draft.Slug, Title: draft.Title, Subtitle: draft.Subtitle, Content: compact.Bytes()}); err != nil {
			return err
		}
		if err := queries.SnapshotDraft(ctx, id); err != nil {
			return err
		}
		return queries.PublishArticle(ctx, id)
	})
}

// Unpublish hides but retains the last snapshot for audit/republish isolation.
func (p *Postgres) Unpublish(ctx context.Context, id, expectedVersion int64) (Draft, error) {
	return p.mutate(ctx, id, expectedVersion, func(queries *q.Queries, _ q.Article) error { return queries.UnpublishArticle(ctx, id) })
}

func (p *Postgres) mutate(ctx context.Context, id, expectedVersion int64, change func(*q.Queries, q.Article) error) (Draft, error) {
	if id <= 0 || expectedVersion <= 0 {
		return Draft{}, ErrInvalid
	}
	tx, err := p.db.BeginTxx(ctx, &sql.TxOptions{Isolation: sql.LevelReadCommitted})
	if err != nil {
		return Draft{}, databaseError(err)
	}
	defer tx.Rollback()
	queries := q.New(tx)
	a, err := queries.LockArticle(ctx, id)
	if err != nil {
		return Draft{}, databaseError(err)
	}
	if a.Version != expectedVersion {
		return Draft{}, ErrConflict
	}
	if err = change(queries, a); err != nil {
		return Draft{}, databaseError(err)
	}
	return commitDraft(ctx, tx, queries, id)
}

func commitDraft(ctx context.Context, tx *sqlx.Tx, queries *q.Queries, id int64) (Draft, error) {
	d, err := queries.GetDraft(ctx, id)
	if err != nil {
		return Draft{}, databaseError(err)
	}
	if err = tx.Commit(); err != nil {
		return Draft{}, databaseError(err)
	}
	return draftFromRow(d), nil
}

func draftFromRow(d q.GetDraftRow) Draft {
	out := Draft{ID: d.ID, Version: d.Version, Title: d.Title, Subtitle: d.Subtitle, Slug: d.Slug, Content: d.Content, Published: d.Published}
	if d.FirstPublishedAt.Valid {
		stamp := d.FirstPublishedAt.Time
		out.FirstPublishedAt = &stamp
	}
	return out
}

func (p *Postgres) PublicDetail(ctx context.Context, slug string) (Publication, error) {
	if !validSlug(slug) {
		return Publication{}, ErrInvalid
	}
	d, err := p.queries.PublicDetail(ctx, slug)
	if err != nil {
		return Publication{}, databaseError(err)
	}
	return Publication{ID: d.ID, Title: d.Title, Subtitle: d.Subtitle, Slug: d.Slug, Content: d.Content, FirstPublishedAt: d.FirstPublishedAt.Time, PublishedAt: d.PublishedAt}, nil
}

// PublicList fetches only summaries using an immutable timestamp/id keyset.
// Zero limit defaults to 20; the maximum is 50. This is not snapshot isolation
// across requests: unpublishing can remove items between pages.
func (p *Postgres) PublicList(ctx context.Context, limit int32, after *PublicCursor) (PublicPage, error) {
	if limit == 0 {
		limit = 20
	}
	if limit < 1 || limit > 50 {
		return PublicPage{}, ErrInvalid
	}
	params := q.PublicListParams{PageLimit: limit + 1}
	if after != nil {
		if after.ID <= 0 || after.FirstPublishedAt.IsZero() || after.FirstPublishedAt.Year() < 1 || after.FirstPublishedAt.Year() > 9999 || after.FirstPublishedAt.Nanosecond()%1000 != 0 {
			return PublicPage{}, ErrInvalid
		}
		params.HasCursor = true
		params.AfterTime = after.FirstPublishedAt
		params.AfterID = after.ID
	}
	rows, err := p.queries.PublicList(ctx, params)
	if err != nil {
		return PublicPage{}, databaseError(err)
	}
	out := PublicPage{Items: make([]PublicationSummary, 0, min(len(rows), int(limit)))}
	if len(rows) > int(limit) {
		rows = rows[:limit]
		last := rows[len(rows)-1]
		out.NextCursor = &PublicCursor{FirstPublishedAt: last.FirstPublishedAt.Time, ID: last.ID}
	}
	for _, d := range rows {
		out.Items = append(out.Items, PublicationSummary{ID: d.ID, Title: d.Title, Subtitle: d.Subtitle, Slug: d.Slug, FirstPublishedAt: d.FirstPublishedAt.Time, PublishedAt: d.PublishedAt})
	}
	return out, nil
}

func databaseError(err error) error {
	if errors.Is(err, sql.ErrNoRows) {
		return ErrNotFound
	}
	var pg *pgconn.PgError
	if errors.As(err, &pg) && pg.Code == "23505" && pg.ConstraintName == "articles_slug_key" {
		return ErrSlugTaken
	}
	return err
}
