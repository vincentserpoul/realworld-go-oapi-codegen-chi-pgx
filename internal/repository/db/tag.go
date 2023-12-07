package db

import (
	"context"
	"fmt"

	"realworld/internal/domain"
)

// implement the interface TagRepository with named args
func (r *Repository) GetTags(ctx context.Context) ([]domain.Tag, error) {
	query := `
		SELECT
			ARRAY_AGG(t.name)
		FROM tag t
	`

	var tags []domain.Tag
	if err := r.pool.QueryRow(ctx, query).Scan(&tags); err != nil {
		return nil, fmt.Errorf("could not scan tag: %w", err)
	}

	return tags, nil
}
