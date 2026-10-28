-- name: CreateArticle :one
INSERT INTO articles (slug) VALUES ($1) RETURNING *;

-- name: CreateDraft :exec
INSERT INTO article_drafts (article_id, title, subtitle, content) VALUES ($1, $2, $3, $4);

-- name: LockArticle :one
SELECT * FROM articles WHERE id = $1 FOR UPDATE;

-- name: GetDraft :one
SELECT a.id, a.slug, a.version, a.published, a.first_published_at, d.title, d.subtitle, d.content
FROM articles a JOIN article_drafts d ON d.article_id = a.id WHERE a.id = $1;

-- name: SaveDraft :exec
UPDATE article_drafts SET title = $2, subtitle = $3, content = $4 WHERE article_id = $1;

-- name: ReviseArticle :exec
UPDATE articles SET slug = $2, version = version + 1, updated_at = now() WHERE id = $1;

-- name: SnapshotDraft :exec
INSERT INTO article_publications (article_id, title, subtitle, content, saved_version)
SELECT d.article_id, d.title, d.subtitle, d.content, a.version FROM article_drafts d
JOIN articles a ON a.id = d.article_id WHERE d.article_id = $1
ON CONFLICT (article_id) DO UPDATE SET title = EXCLUDED.title, subtitle = EXCLUDED.subtitle,
content = EXCLUDED.content, saved_version = EXCLUDED.saved_version, published_at = now();

-- name: LockDraftForPublication :one
SELECT a.slug, d.title, d.subtitle, d.content FROM article_drafts d
JOIN articles a ON a.id = d.article_id WHERE d.article_id = $1 FOR UPDATE OF d;

-- name: PublishArticle :exec
UPDATE articles SET published = true, first_published_at = COALESCE(first_published_at, now()),
version = version + 1, updated_at = now() WHERE id = $1;

-- name: UnpublishArticle :exec
UPDATE articles SET published = false, version = version + 1, updated_at = now() WHERE id = $1;

-- name: PublicDetail :one
SELECT a.id, a.slug, a.first_published_at, p.title, p.subtitle, p.content, p.published_at
FROM articles a JOIN article_publications p ON p.article_id = a.id WHERE a.slug = $1 AND a.published;

-- name: PublicList :many
SELECT a.id, a.slug, a.first_published_at, p.title, p.subtitle, p.published_at
FROM articles a JOIN article_publications p ON p.article_id = a.id WHERE a.published
AND (NOT sqlc.arg(has_cursor)::boolean OR (a.first_published_at, a.id) < (sqlc.arg(after_time)::timestamptz, sqlc.arg(after_id)::bigint))
ORDER BY a.first_published_at DESC, a.id DESC LIMIT sqlc.arg(page_limit);
