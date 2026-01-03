package domain

import (
	"context"
	"time"

	"github.com/google/uuid"
)

type User struct {
	ID        uuid.UUID `db:"id" json:"id"`
	Email     string    `db:"email" json:"email"`
	Username  string    `db:"username" json:"username"`
	Password  string    `db:"pwd" json:"pwd"`
	Bio       string    `db:"bio" json:"bio"`
	Image     string    `db:"img" json:"img"`
	CreatedAt time.Time `db:"created_at" json:"created_at"`
	UpdatedAt time.Time `db:"updated_at" json:"updated_at"`
}

//nolint:iface //for extension
type UserService interface {
	UserRepository
}

//nolint:iface //for extension
type UserRepository interface {
	RegisterUser(
		ctx context.Context,
		userID uuid.UUID,
		email, username, password string,
	) (*User, error)
	AuthUser(ctx context.Context, email, password string) (*User, string, error)
	GetUser(ctx context.Context, username string) (*User, error)
	GetCurrentUser(ctx context.Context, userID uuid.UUID) (*User, error)
	UpdateUser(
		ctx context.Context,
		userID uuid.UUID,
		username, email, password, bio, image *string,
	) (*User, error)
}

type Profile struct {
	Username  string `db:"username" json:"username"`
	Bio       string `db:"bio" json:"bio"`
	Image     string `db:"img" json:"img"`
	Following bool   `db:"following" json:"following"`
}

//nolint:iface //for extension
type ProfileService interface {
	ProfileRepository
}

//nolint:iface //for extension
type ProfileRepository interface {
	GetProfile(ctx context.Context, userID uuid.UUID, username string) (*Profile, error)
	FollowUser(ctx context.Context, userID uuid.UUID, followUsername string) (*Profile, error)
	UnfollowUser(ctx context.Context, userID uuid.UUID, unfollowUsername string) (*Profile, error)
}
