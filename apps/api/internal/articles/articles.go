// Package articles defines the bounded article persistence use cases. It has no
// HTTP/auth policy: callers must authorize access to every draft/mutation.
package articles

import (
	"context"
	"encoding/json"
	"errors"
	"time"
)

var (
	ErrInvalid   = errors.New("invalid article input")
	ErrNotFound  = errors.New("article not found")
	ErrConflict  = errors.New("article version conflict")
	ErrSlugTaken = errors.New("article slug already taken")
)

type Input struct {
	Title    string          `json:"title"`
	Subtitle string          `json:"subtitle"`
	Slug     string          `json:"slug"`
	Content  json.RawMessage `json:"content"`
}

// Version is an aggregate revision, not a draft-only counter. Every successful
// save, publish and unpublish increments it, including repeated state changes.
type Draft struct {
	ID               int64           `json:"id"`
	Version          int64           `json:"version"`
	Title            string          `json:"title"`
	Subtitle         string          `json:"subtitle"`
	Slug             string          `json:"slug"`
	Content          json.RawMessage `json:"content"`
	FirstPublishedAt *time.Time      `json:"firstPublishedAt"`
	Published        bool            `json:"published"`
}

type Publication struct {
	ID               int64           `json:"id"`
	Title            string          `json:"title"`
	Subtitle         string          `json:"subtitle"`
	Slug             string          `json:"slug"`
	Content          json.RawMessage `json:"content"`
	FirstPublishedAt time.Time       `json:"firstPublishedAt"`
	PublishedAt      time.Time       `json:"publishedAt"`
}

// PublicationSummary deliberately excludes the potentially large document.
type PublicationSummary struct {
	ID               int64     `json:"id"`
	Title            string    `json:"title"`
	Subtitle         string    `json:"subtitle"`
	Slug             string    `json:"slug"`
	FirstPublishedAt time.Time `json:"firstPublishedAt"`
	PublishedAt      time.Time `json:"publishedAt"`
}

// PublicCursor is a typed keyset position. HTTP cursor encoding/decoding belongs
// to the transport boundary; never interpolate it into SQL.
type PublicCursor struct {
	FirstPublishedAt time.Time `json:"firstPublishedAt"`
	ID               int64     `json:"id"`
}

type PublicPage struct {
	Items      []PublicationSummary `json:"items"`
	NextCursor *PublicCursor        `json:"nextCursor"`
}

// Repository is the application port. Publish snapshots the exact saved draft
// selected by expectedVersion; unsaved content is deliberately not accepted.
type Repository interface {
	Create(context.Context, Input) (Draft, error)
	GetDraft(context.Context, int64) (Draft, error)
	Save(context.Context, int64, int64, Input) (Draft, error)
	Publish(context.Context, int64, int64) (Draft, error)
	Unpublish(context.Context, int64, int64) (Draft, error)
	PublicDetail(context.Context, string) (Publication, error)
	PublicList(context.Context, int32, *PublicCursor) (PublicPage, error)
}
