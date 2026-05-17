package domain

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/induzo/gocom/http/health"
)

// enforce service interface
var _ APIService = (*APISvc)(nil)

type APISvc struct {
	repository APIRepository
}

func NewAPISvc(repo APIRepository) *APISvc {
	return &APISvc{
		repository: repo,
	}
}

func (as *APISvc) GetTags(ctx context.Context) ([]Tag, error) {
	tags, err := as.repository.GetTags(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get tags: %w", err)
	}

	return tags, nil
}

func (as *APISvc) GetArticles(
	ctx context.Context,
	userID uuid.UUID,
	author, tag, favorited *string,
	limit, offset *int,
) ([]*Article, error) {
	articles, err := as.repository.GetArticles(ctx, userID, author, tag, favorited, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("failed to get articles: %w", err)
	}

	return articles, nil
}

func (as *APISvc) GetArticle(ctx context.Context, userID uuid.UUID, slug string) (*Article, error) {
	article, err := as.repository.GetArticle(ctx, userID, slug)
	if err != nil {
		return nil, fmt.Errorf("failed to get article: %w", err)
	}

	return article, nil
}

func (as *APISvc) GetFeedArticles(
	ctx context.Context,
	userID uuid.UUID,
	username, tag, favorited *string,
	limit, offset *int,
) ([]*Article, error) {
	articles, err := as.repository.GetFeedArticles(
		ctx,
		userID,
		username,
		tag,
		favorited,
		limit,
		offset,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to get feed articles: %w", err)
	}

	return articles, nil
}

func (as *APISvc) CreateArticle(
	ctx context.Context,
	userID uuid.UUID,
	title, description, body string,
	tagList []string,
) (*Article, error) {
	article, err := as.repository.CreateArticle(ctx, userID, title, description, body, tagList)
	if err != nil {
		return nil, fmt.Errorf("failed to create article: %w", err)
	}

	return article, nil
}

func (as *APISvc) UpdateArticle(
	ctx context.Context,
	userID uuid.UUID,
	slug string,
	title, description, body *string,
) (*Article, error) {
	article, err := as.repository.UpdateArticle(ctx, userID, slug, title, description, body)
	if err != nil {
		return nil, fmt.Errorf("failed to update article: %w", err)
	}

	return article, nil
}

func (as *APISvc) DeleteArticle(ctx context.Context, userID uuid.UUID, slug string) error {
	if err := as.repository.DeleteArticle(ctx, userID, slug); err != nil {
		return fmt.Errorf("failed to delete article: %w", err)
	}

	return nil
}

func (as *APISvc) FavoriteArticle(
	ctx context.Context,
	userID uuid.UUID,
	slug string,
) (*Article, error) {
	article, err := as.repository.FavoriteArticle(ctx, userID, slug)
	if err != nil {
		return nil, fmt.Errorf("failed to favorite article: %w", err)
	}

	return article, nil
}

func (as *APISvc) UnfavoriteArticle(
	ctx context.Context,
	userID uuid.UUID,
	slug string,
) (*Article, error) {
	article, err := as.repository.UnfavoriteArticle(ctx, userID, slug)
	if err != nil {
		return nil, fmt.Errorf("failed to unfavorite article: %w", err)
	}

	return article, nil
}

func (as *APISvc) RegisterUser(
	ctx context.Context,
	userID uuid.UUID,
	email, username, password string,
) (*User, error) {
	user, err := as.repository.RegisterUser(ctx, userID, email, username, password)
	if err != nil {
		return nil, fmt.Errorf("failed to register user: %w", err)
	}

	return user, nil
}

func (as *APISvc) AuthUser(ctx context.Context, email, password string) (*User, string, error) {
	user, token, err := as.repository.AuthUser(ctx, email, password)
	if err != nil {
		return nil, "", fmt.Errorf("failed to authenticate user: %w", err)
	}

	return user, token, nil
}

func (as *APISvc) GetUser(ctx context.Context, username string) (*User, error) {
	user, err := as.repository.GetUser(ctx, username)
	if err != nil {
		return nil, fmt.Errorf("failed to get user: %w", err)
	}

	return user, nil
}

func (as *APISvc) GetCurrentUser(ctx context.Context, userID uuid.UUID) (*User, error) {
	user, err := as.repository.GetCurrentUser(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to get current user: %w", err)
	}

	return user, nil
}

func (as *APISvc) UpdateUser(
	ctx context.Context,
	userID uuid.UUID,
	username, email, password, bio, image *string,
) (*User, error) {
	user, err := as.repository.UpdateUser(ctx, userID, username, email, password, bio, image)
	if err != nil {
		return nil, fmt.Errorf("failed to update user: %w", err)
	}

	return user, nil
}

func (as *APISvc) GetProfile(
	ctx context.Context,
	userID uuid.UUID,
	username string,
) (*Profile, error) {
	profile, err := as.repository.GetProfile(ctx, userID, username)
	if err != nil {
		return nil, fmt.Errorf("failed to get profile: %w", err)
	}

	return profile, nil
}

func (as *APISvc) FollowUser(
	ctx context.Context,
	userID uuid.UUID,
	followUsername string,
) (*Profile, error) {
	profile, err := as.repository.FollowUser(ctx, userID, followUsername)
	if err != nil {
		return nil, fmt.Errorf("failed to follow user: %w", err)
	}

	return profile, nil
}

func (as *APISvc) UnfollowUser(
	ctx context.Context,
	userID uuid.UUID,
	unfollowUsername string,
) (*Profile, error) {
	profile, err := as.repository.UnfollowUser(ctx, userID, unfollowUsername)
	if err != nil {
		return nil, fmt.Errorf("failed to unfollow user: %w", err)
	}

	return profile, nil
}

func (as *APISvc) GetComments(
	ctx context.Context,
	userID uuid.UUID,
	slug string,
) ([]*Comment, error) {
	comments, err := as.repository.GetComments(ctx, userID, slug)
	if err != nil {
		return nil, fmt.Errorf("failed to get comments: %w", err)
	}

	return comments, nil
}

func (as *APISvc) AddComment(
	ctx context.Context,
	authorID uuid.UUID,
	slug, body string,
) (*Comment, error) {
	comment, err := as.repository.AddComment(ctx, authorID, slug, body)
	if err != nil {
		return nil, fmt.Errorf("failed to add comment: %w", err)
	}

	return comment, nil
}

func (as *APISvc) DeleteComment(ctx context.Context, slug string, id int) error {
	if err := as.repository.DeleteComment(ctx, slug, id); err != nil {
		return fmt.Errorf("failed to delete comment: %w", err)
	}

	return nil
}

func (as *APISvc) GetShutdownFuncs() map[string]func(ctx context.Context) error {
	return as.repository.GetShutdownFuncs()
}

func (as *APISvc) GetHealthChecks() []health.CheckConfig {
	return as.repository.GetHealthChecks()
}
