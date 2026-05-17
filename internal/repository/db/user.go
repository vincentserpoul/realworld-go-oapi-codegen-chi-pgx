package db

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"

	"realworld/internal/domain"
)

func (r *Repository) RegisterUser(
	ctx context.Context,
	userID uuid.UUID,
	username,
	email,
	password string,
) (*domain.User, error) {
	rows, err := r.pool.Query(ctx, `
        INSERT INTO appuser (id, username, email, pwd)
        VALUES (@userID, @username, @email, @password)
        RETURNING id, email, username, pwd, bio, img, created_at, updated_at`,
		pgx.NamedArgs{
			"userID":   userID,
			"username": username,
			"email":    email,
			"password": password,
		},
	)
	if err != nil {
		return nil, fmt.Errorf("could not insert user: %w", err)
	}

	user, errA := pgx.CollectExactlyOneRow(rows, pgx.RowToAddrOfStructByName[domain.User])
	if errA != nil {
		return nil, fmt.Errorf("could not collect rows: %w", errA)
	}

	return user, nil
}

func (r *Repository) AuthUser(
	ctx context.Context,
	email, password string,
) (*domain.User, string, error) {
	rows, err := r.pool.Query(ctx, `
		SELECT id, username, pwd, email, bio, img, created_at, updated_at
		FROM appuser
		WHERE email = @email AND pwd = @password`,
		pgx.NamedArgs{
			"email":    email,
			"password": password,
		},
	)
	if err != nil {
		return nil, "", fmt.Errorf("could not get user: %w", err)
	}

	usr, errA := pgx.CollectExactlyOneRow(rows, pgx.RowToAddrOfStructByName[domain.User])
	if errA != nil {
		return nil, "", fmt.Errorf("could not collect rows: %w", errA)
	}

	token := "123"

	return usr, token, nil
}

func (r *Repository) GetUser(ctx context.Context, username string) (*domain.User, error) {
	rows, err := r.pool.Query(ctx, `
		SELECT id, username, email, pwd, bio, img
		FROM appuser
		WHERE username = @username`,
		pgx.NamedArgs{
			"username": username,
		},
	)
	if err != nil {
		return nil, fmt.Errorf("could not get user: %w", err)
	}

	user, errA := pgx.CollectExactlyOneRow(rows, pgx.RowToAddrOfStructByName[domain.User])
	if errA != nil {
		return nil, fmt.Errorf("could not collect rows: %w", errA)
	}

	return user, nil
}

func (r *Repository) GetCurrentUser(ctx context.Context, userID uuid.UUID) (*domain.User, error) {
	rows, err := r.pool.Query(ctx, `
		SELECT id, username, email, pwd, bio, img, created_at, updated_at
		FROM appuser
		WHERE id = @userID`,
		pgx.NamedArgs{
			"userID": userID,
		},
	)
	if err != nil {
		return nil, fmt.Errorf("could not get user: %w", err)
	}

	user, errA := pgx.CollectExactlyOneRow(rows, pgx.RowToAddrOfStructByName[domain.User])
	if errA != nil {
		return nil, fmt.Errorf("could not collect rows: %w", errA)
	}

	return user, nil
}

var ErrNoFieldsToUpdate = errors.New("no fields to update")

func (r *Repository) UpdateUser(
	ctx context.Context,
	userID uuid.UUID,
	username, email, password, bio, image *string,
) (*domain.User, error) {
	args := pgx.NamedArgs{
		"id": userID,
	}
	updatedFields := []string{}

	if username != nil {
		updatedFields = append(updatedFields, "username = @username")
		args["username"] = username
	}

	if email != nil {
		updatedFields = append(updatedFields, "email = @email")
		args["email"] = email
	}

	if password != nil {
		updatedFields = append(updatedFields, "pwd = @password")
		args["password"] = password
	}

	if bio != nil {
		updatedFields = append(updatedFields, "bio = @bio")
		args["bio"] = bio
	}

	if image != nil {
		updatedFields = append(updatedFields, "img = @image")
		args["image"] = image
	}

	if len(updatedFields) == 0 {
		return nil, fmt.Errorf("UpdateUser: %w", ErrNoFieldsToUpdate)
	}

	query := `
	UPDATE appuser
	SET ` + strings.Join(updatedFields, `, `) + `
	WHERE id = @id
	RETURNING id, username, email, pwd, bio, img, created_at, updated_at`

	rows, err := r.pool.Query(ctx, query, args)
	if err != nil {
		return nil, fmt.Errorf("could not update user: %w", err)
	}

	user, errA := pgx.CollectExactlyOneRow(rows, pgx.RowToAddrOfStructByName[domain.User])
	if errA != nil {
		return nil, fmt.Errorf("could not collect rows: %w", errA)
	}

	return user, nil
}
