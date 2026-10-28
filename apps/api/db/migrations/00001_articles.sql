-- +goose Up
CREATE TABLE admins (
    id bigint GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    email text NOT NULL UNIQUE CHECK (length(email) BETWEEN 3 AND 254 AND email = lower(email)),
    password_hash text NOT NULL CHECK (length(password_hash) BETWEEN 20 AND 1024),
    created_at timestamptz NOT NULL DEFAULT now()
);

CREATE TABLE articles (
    id bigint GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    slug varchar(120) COLLATE "C" NOT NULL UNIQUE CHECK (slug ~ '^[a-z0-9]+(-[a-z0-9]+)*$'),
    version bigint NOT NULL DEFAULT 1 CHECK (version > 0),
    published boolean NOT NULL DEFAULT false,
    first_published_at timestamptz,
    created_at timestamptz NOT NULL DEFAULT now(),
    updated_at timestamptz NOT NULL DEFAULT now(),
    CHECK (NOT published OR first_published_at IS NOT NULL)
);

CREATE TABLE article_drafts (
    article_id bigint PRIMARY KEY REFERENCES articles(id) ON DELETE CASCADE,
    title text NOT NULL CHECK (length(btrim(title)) BETWEEN 1 AND 200),
    subtitle text NOT NULL DEFAULT '' CHECK (length(subtitle) <= 500),
    content jsonb NOT NULL CHECK (jsonb_typeof(content) = 'object' AND content @> '{"schemaVersion":1}' AND content ? 'doc' AND jsonb_typeof(content->'doc') = 'object')
);

CREATE TABLE article_publications (
    article_id bigint PRIMARY KEY REFERENCES articles(id) ON DELETE CASCADE,
    title text NOT NULL CHECK (length(btrim(title)) BETWEEN 1 AND 200),
    subtitle text NOT NULL CHECK (length(subtitle) <= 500),
    content jsonb NOT NULL CHECK (jsonb_typeof(content) = 'object' AND content @> '{"schemaVersion":1}' AND content ? 'doc' AND jsonb_typeof(content->'doc') = 'object'),
    saved_version bigint NOT NULL CHECK (saved_version > 0),
    published_at timestamptz NOT NULL DEFAULT now()
);

CREATE INDEX articles_public_order ON articles (first_published_at DESC, id DESC) WHERE published;

CREATE TABLE media (
    id bigint GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    article_id bigint NOT NULL REFERENCES articles(id) ON DELETE CASCADE,
    object_key text NOT NULL UNIQUE CHECK (length(object_key) BETWEEN 1 AND 512),
    mime_type text NOT NULL CHECK (mime_type IN ('image/jpeg', 'image/png', 'image/webp')),
    size_bytes bigint NOT NULL CHECK (size_bytes BETWEEN 1 AND 10485760),
    width integer NOT NULL CHECK (width BETWEEN 1 AND 20000),
    height integer NOT NULL CHECK (height BETWEEN 1 AND 20000),
    alt_text text NOT NULL DEFAULT '' CHECK (length(alt_text) <= 1000),
    created_at timestamptz NOT NULL DEFAULT now()
);
CREATE INDEX media_article_id ON media(article_id);

-- +goose StatementBegin
CREATE FUNCTION protect_article_identity() RETURNS trigger LANGUAGE plpgsql AS $$
BEGIN
    IF OLD.first_published_at IS NOT NULL AND
       (NEW.slug IS DISTINCT FROM OLD.slug OR NEW.first_published_at IS DISTINCT FROM OLD.first_published_at) THEN
        RAISE EXCEPTION 'published article identity is immutable' USING ERRCODE = '23514';
    END IF;
    RETURN NEW;
END;
$$;
-- +goose StatementEnd
CREATE TRIGGER protect_article_identity BEFORE UPDATE ON articles FOR EACH ROW EXECUTE FUNCTION protect_article_identity();

-- +goose Down
DROP TABLE media;
DROP TABLE article_publications;
DROP TABLE article_drafts;
DROP TABLE articles;
DROP FUNCTION protect_article_identity();
DROP TABLE admins;
