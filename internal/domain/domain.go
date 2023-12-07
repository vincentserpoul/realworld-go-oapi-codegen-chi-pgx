package domain

import (
	"context"

	"github.com/induzo/gocom/http/health"
)

type APIService interface {
	APIRepository
	GetShutdownFuncs() map[string]func(ctx context.Context) error
	GetHealthChecks() []health.CheckConfig
}

type APIRepository interface {
	ArticleRepository
	ProfileRepository
	TagRepository
	UserRepository
	CommentRepository
	GetShutdownFuncs() map[string]func(ctx context.Context) error
	GetHealthChecks() []health.CheckConfig
}
