package db

import (
	"context"
	"fmt"

	"github.com/gofrs/uuid/v5"
	"github.com/jackc/pgx/v5"

	"realworld/internal/domain"
)

// implement the interface ProfileRepository with named args
func (r *Repository) GetProfile(
	ctx context.Context,
	userID uuid.UUID,
	username string,
) (*domain.Profile, error) {
	// query with named args
	query := `
		SELECT
			u.username,
			u.bio,
			u.img,
			EXISTS (
				SELECT 1
				FROM appuser_follows f
				WHERE f.followee_id = u.id
				AND f.follower_id = @userID
			) AS following
		FROM appuser u
		WHERE u.username = @username
	`

	// named parameters
	args := pgx.NamedArgs{
		"userID":   userID,
		"username": username,
	}

	rows, errR := r.pool.Query(ctx, query, args)
	if errR != nil {
		return nil, fmt.Errorf("could not get profile: %w", errR)
	}

	profile, errA := pgx.CollectExactlyOneRow(rows, pgx.RowToAddrOfStructByName[domain.Profile])
	if errA != nil {
		return nil, fmt.Errorf("could not collect row: %w", errA)
	}

	return profile, nil
}

// implement the interface ProfileRepository with named args
func (r *Repository) FollowUser(
	ctx context.Context,
	followerID uuid.UUID,
	username string,
) (*domain.Profile, error) {
	// query with named args
	query := `
		WITH profile AS (
			SELECT
				u.id,
				u.username,
				u.bio,
				u.img
			FROM appuser u
			WHERE u.username = @username
		)
		INSERT INTO appuser_follows (followee_id, follower_id)
		SELECT
			p.id,
			@followerID
		FROM profile p
		ON CONFLICT DO NOTHING
		RETURNING
			(SELECT username FROM profile),
			(SELECT bio FROM profile),
			(SELECT img FROM profile),
			true AS following
	`

	// named parameters
	args := pgx.NamedArgs{
		"followerID": followerID,
		"username":   username,
	}

	rows, errR := r.pool.Query(ctx, query, args)
	if errR != nil {
		return nil, fmt.Errorf("could not follow profile: %w", errR)
	}

	profile, errA := pgx.CollectExactlyOneRow(rows, pgx.RowToAddrOfStructByName[domain.Profile])
	if errA != nil {
		return nil, fmt.Errorf("could not collect row: %w", errA)
	}

	return profile, nil
}

// implement the interface ProfileRepository with named args
func (r *Repository) UnfollowUser(
	ctx context.Context,
	followerID uuid.UUID,
	username string,
) (*domain.Profile, error) {
	// query with named args
	query := `
		WITH profile AS (
			SELECT
				u.id,
				u.username,
				u.bio,
				u.img
			FROM appuser u
			WHERE u.username = @username
		)
		DELETE FROM appuser_follows f
		USING profile p
		WHERE f.followee_id = p.id
		AND f.follower_id = @followerID
		RETURNING
			p.username,
			p.bio,
			p.img,
			false as following
	`

	// named parameters
	args := pgx.NamedArgs{
		"followerID": followerID,
		"username":   username,
	}

	rows, errR := r.pool.Query(ctx, query, args)
	if errR != nil {
		return nil, fmt.Errorf("could not unfollow profile: %w", errR)
	}

	profile, errA := pgx.CollectExactlyOneRow(rows, pgx.RowToAddrOfStructByName[domain.Profile])
	if errA != nil {
		return nil, fmt.Errorf("could not collect row: %w", errA)
	}

	return profile, nil
}
