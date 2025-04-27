package domain

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/gosimple/slug"
)

type Tag string

//nolint:iface //for extension
type TagService interface {
	GetTags(ctx context.Context) ([]Tag, error)
}

//nolint:iface //for extension
type TagRepository interface {
	GetTags(ctx context.Context) ([]Tag, error)
}

type Article struct {
	ID             uuid.UUID `db:"id" json:"id"`
	Slug           string    `db:"slug" json:"slug"`
	Title          string    `db:"title" json:"title"`
	Description    string    `db:"description" json:"description"`
	Body           string    `db:"body" json:"body"`
	TagList        []Tag     `db:"tag_list" json:"tag_list"`
	CreatedAt      time.Time `db:"created_at" json:"created_at"`
	UpdatedAt      time.Time `db:"updated_at" json:"updated_at"`
	Favorited      bool      `db:"favorited" json:"favorited"`
	FavoritesCount int       `db:"favorites_count" json:"favorites_count"`
	Author         Profile   `db:"author" json:"author"`
}

func GetSlugFromTitle(title string) string {
	return slug.Make(title)
}

//nolint:iface //for extension
type ArticleService interface {
	ArticleRepository
}

//nolint:iface //for extension
type ArticleRepository interface {
	GetArticles(
		ctx context.Context,
		userID uuid.UUID,
		author, tag, favorited *string,
		limit, offset *int,
	) ([]*Article, error)
	GetArticle(ctx context.Context, userID uuid.UUID, slug string) (*Article, error)
	GetFeedArticles(
		ctx context.Context,
		userID uuid.UUID,
		username, tag, favorited *string,
		limit, offset *int,
	) ([]*Article, error)
	CreateArticle(
		ctx context.Context,
		userID uuid.UUID,
		title, description, body string,
		tagList []string,
	) (*Article, error)
	UpdateArticle(ctx context.Context, userID uuid.UUID, slug string, title, description, body *string) (*Article, error)
	DeleteArticle(ctx context.Context, userID uuid.UUID, slug string) error
	FavoriteArticle(ctx context.Context, userID uuid.UUID, slug string) (*Article, error)
	UnfavoriteArticle(ctx context.Context, userID uuid.UUID, slug string) (*Article, error)
}
