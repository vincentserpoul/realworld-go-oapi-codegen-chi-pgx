package domain

import (
	"context"

	"github.com/induzo/gocom/http/health"
)

//nolint:iface //for extension
type APIService interface {
	APIRepository
	GetShutdownFuncs() map[string]func(ctx context.Context) error
	GetHealthChecks() []health.CheckConfig
}

//nolint:iface //for extension
type APIRepository interface {
	ArticleRepository
	ProfileRepository
	TagRepository
	UserRepository
	CommentRepository
	GetShutdownFuncs() map[string]func(ctx context.Context) error
	GetHealthChecks() []health.CheckConfig
}
