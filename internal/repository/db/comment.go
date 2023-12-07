package db

import (
	"context"
	"fmt"

	"github.com/gofrs/uuid/v5"
	"github.com/induzo/gocom/database/pginit/v2"
	"github.com/jackc/pgx/v5"

	"realworld/internal/domain"
)

func (r *Repository) GetComments(
	ctx context.Context,
	userID uuid.UUID,
	slug string,
) ([]*domain.Comment, error) {
	// query with named args
	query := `
		SELECT
			JSON_BUILD_OBJECT(
				'id', c.id,
				'body', c.body,
				'created_at', c.created_at,
				'updated_at', c.updated_at,
				'author', (
					SELECT JSON_BUILD_OBJECT(
						'username', u.username,
						'bio', u.bio,
						'img', u.img,
						'following', EXISTS(
							SELECT 1
							FROM appuser_follows
							WHERE follower_id = @userID
							AND followee_id = u.id
						)
					)
					FROM appuser u
					WHERE u.id = c.author_id
				)
			)
		FROM comment c
		JOIN article a ON c.article_id = a.id AND a.slug = @slug
	`

	// named parameters
	args := pgx.NamedArgs{
		"slug":   slug,
		"userID": userID,
	}

	rows, errR := r.pool.Query(ctx, query, args)
	if errR != nil {
		return nil, fmt.Errorf("could not get profile: %w", errR)
	}

	comments, errA := pgx.CollectRows(rows, pginit.JSONRowToAddrOfStruct[domain.Comment])
	if errA != nil {
		return nil, fmt.Errorf("could not collect row: %w", errA)
	}

	return comments, nil
}

func (r *Repository) AddComment(
	ctx context.Context,
	authorID uuid.UUID,
	slug, body string,
) (*domain.Comment, error) {
	// query with named args
	query := `
		INSERT INTO comment (body, author_id, article_id)
		VALUES (
			@body,
			@authorID,
			(SELECT id FROM article WHERE slug = @slug)
		)
		RETURNING
			JSON_BUILD_OBJECT(
				'id', id,
				'body', body,
				'created_at', created_at,
				'updated_at', updated_at,
				'author', (
					SELECT JSON_BUILD_OBJECT(
						'username', u.username,
						'bio', u.bio,
						'img', u.img,
						'following', EXISTS(
							SELECT 1
							FROM appuser_follows
							WHERE follower_id = @userID
							AND followee_id = u.id
						)
					)
					FROM appuser u
					WHERE u.id = author_id
				)
			)
	`

	// named parameters
	args := pgx.NamedArgs{
		"slug":     slug,
		"body":     body,
		"authorID": authorID,
	}

	rows, err := r.pool.Query(ctx, query, args)
	if err != nil {
		return nil, fmt.Errorf("could not insert comment: %w", err)
	}

	comment, errA := pgx.CollectExactlyOneRow(rows, pginit.JSONRowToAddrOfStruct[domain.Comment])
	if errA != nil {
		return nil, fmt.Errorf("could not insert comment: %w", errA)
	}

	return comment, nil
}

func (r *Repository) DeleteComment(
	ctx context.Context,
	slug string, commentID int,
) error {
	// query with named args
	query := `
		DELETE FROM comment c
		WHERE c.id = @commentID
		AND c.article_id = (SELECT id FROM article WHERE slug = @slug)
	`

	// named parameters
	args := pgx.NamedArgs{
		"slug": slug,
		"id":   commentID,
	}

	_, err := r.pool.Exec(ctx, query, args)
	if err != nil {
		return fmt.Errorf("could not delete comment: %w", err)
	}

	return nil
}
