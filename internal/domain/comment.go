package domain

import (
	"context"
	"time"

	"github.com/google/uuid"
)

type Comment struct {
	ID        int       `db:"id" json:"id"`
	Body      string    `db:"body" json:"body"`
	CreatedAt time.Time `db:"created_at" json:"created_at"`
	UpdatedAt time.Time `db:"updated_at" json:"updated_at"`
	Author    Profile   `db:"author" json:"author"`
}

type CommentService interface {
	CommentRepository
}

type CommentRepository interface {
	GetComments(ctx context.Context, userID uuid.UUID, slug string) ([]*Comment, error)
	AddComment(ctx context.Context, authorID uuid.UUID, slug, body string) (*Comment, error)
	DeleteComment(ctx context.Context, slug string, id int) error
}
