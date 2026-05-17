package httpapi

import (
	"context"
	"fmt"
	"time"

	"github.com/go-chi/jwtauth/v5"
	"github.com/google/uuid"
	"github.com/lestrrat-go/jwx/v2/jwt"

	"realworld/internal/domain"
)

// make sure our struct implement the interface
var _ StrictServerInterface = &StrictAPIServer{}

// StrictServerInterface represents all server handlers.
type StrictAPIServer struct {
	svc       domain.APIService
	tokenAuth *jwtauth.JWTAuth
}

func NewStrictAPIServer(svc domain.APIService, tokenAuth *jwtauth.JWTAuth) *StrictAPIServer {
	return &StrictAPIServer{
		svc:       svc,
		tokenAuth: tokenAuth,
	}
}

// Get recent articles globally
// (GET /articles)
func (s *StrictAPIServer) GetArticles(
	ctx context.Context,
	request GetArticlesRequestObject,
) (GetArticlesResponseObject, error) {
	// get the request context
	articles, errA := s.svc.GetArticles(
		ctx,
		getUserIDFromContext(ctx),
		request.Params.Author,
		request.Params.Tag,
		request.Params.Favorited,
		request.Params.Limit,
		request.Params.Offset,
	)
	if errA != nil {
		return GetArticles422JSONResponse{}, fmt.Errorf("get articles: %w", errA)
	}

	return GetArticles200JSONResponse{
		MultipleArticlesResponseJSONResponse: MultipleArticlesResponseJSONResponse{
			Articles:      fromDomainArticles(articles),
			ArticlesCount: len(articles),
		},
	}, nil
}

// Create an article
// (POST /articles)
func (s *StrictAPIServer) CreateArticle(
	ctx context.Context,
	request CreateArticleRequestObject,
) (CreateArticleResponseObject, error) {
	art, err := s.svc.CreateArticle(
		ctx,
		getUserIDFromContext(ctx),
		request.Body.Article.Title,
		request.Body.Article.Description,
		request.Body.Article.Body,
		*request.Body.Article.TagList,
	)
	if err != nil {
		return CreateArticle422JSONResponse{}, fmt.Errorf("create article: %w", err)
	}

	return CreateArticle201JSONResponse{
		SingleArticleResponseJSONResponse: SingleArticleResponseJSONResponse{
			Article: fromDomainArticle(art),
		},
	}, nil
}

// Get recent articles from users you follow
// (GET /articles/feed)
func (s *StrictAPIServer) GetArticlesFeed(
	ctx context.Context,
	request GetArticlesFeedRequestObject,
) (GetArticlesFeedResponseObject, error) {
	articles, errA := s.svc.GetFeedArticles(
		ctx,
		getUserIDFromContext(ctx),
		nil,
		nil,
		nil,
		request.Params.Limit,
		request.Params.Offset,
	)
	if errA != nil {
		return GetArticlesFeed422JSONResponse{}, fmt.Errorf("get articles feed: %w", errA)
	}

	return GetArticlesFeed200JSONResponse{
		MultipleArticlesResponseJSONResponse: MultipleArticlesResponseJSONResponse{
			Articles:      fromDomainArticles(articles),
			ArticlesCount: len(articles),
		},
	}, nil
}

// Delete an article
// (DELETE /articles/{slug})
func (s *StrictAPIServer) DeleteArticle(
	ctx context.Context,
	request DeleteArticleRequestObject,
) (DeleteArticleResponseObject, error) {
	if err := s.svc.DeleteArticle(ctx, getUserIDFromContext(ctx), request.Slug); err != nil {
		return DeleteArticle422JSONResponse{}, fmt.Errorf("delete article: %w", err)
	}

	return DeleteArticle200Response{}, nil
}

// Get an article
// (GET /articles/{slug})
func (s *StrictAPIServer) GetArticle(
	ctx context.Context,
	request GetArticleRequestObject,
) (GetArticleResponseObject, error) {
	article, errA := s.svc.GetArticle(ctx, getUserIDFromContext(ctx), request.Slug)
	if errA != nil {
		return GetArticle422JSONResponse{}, fmt.Errorf("get article: %w", errA)
	}

	return GetArticle200JSONResponse{
		SingleArticleResponseJSONResponse: SingleArticleResponseJSONResponse{
			Article: fromDomainArticle(article),
		},
	}, nil
}

// Update an article
// (PUT /articles/{slug})
func (s *StrictAPIServer) UpdateArticle(
	ctx context.Context,
	request UpdateArticleRequestObject,
) (UpdateArticleResponseObject, error) {
	art, err := s.svc.UpdateArticle(
		ctx,
		getUserIDFromContext(ctx),
		request.Slug,
		request.Body.Article.Title,
		request.Body.Article.Description,
		request.Body.Article.Body,
	)
	if err != nil {
		return UpdateArticle422JSONResponse{}, fmt.Errorf("update article: %w", err)
	}

	return UpdateArticle200JSONResponse{
		SingleArticleResponseJSONResponse: SingleArticleResponseJSONResponse{
			Article: fromDomainArticle(art),
		},
	}, nil
}

// Get comments for an article
// (GET /articles/{slug}/comments)
func (s *StrictAPIServer) GetArticleComments(
	ctx context.Context,
	request GetArticleCommentsRequestObject,
) (GetArticleCommentsResponseObject, error) {
	comments, err := s.svc.GetComments(ctx, getUserIDFromContext(ctx), request.Slug)
	if err != nil {
		return GetArticleComments422JSONResponse{}, fmt.Errorf("get article comment: %w", err)
	}

	return GetArticleComments200JSONResponse{
		MultipleCommentsResponseJSONResponse: MultipleCommentsResponseJSONResponse{
			Comments: fromDomainComments(comments),
		},
	}, nil
}

// Create a comment for an article
// (POST /articles/{slug}/comments)
func (s *StrictAPIServer) CreateArticleComment(
	ctx context.Context,
	request CreateArticleCommentRequestObject,
) (CreateArticleCommentResponseObject, error) {
	comment, err := s.svc.AddComment(
		ctx,
		getUserIDFromContext(ctx),
		request.Slug,
		request.Body.Comment.Body,
	)
	if err != nil {
		return CreateArticleComment422JSONResponse{}, fmt.Errorf("create article comment: %w", err)
	}

	return CreateArticleComment200JSONResponse{
		SingleCommentResponseJSONResponse: SingleCommentResponseJSONResponse{
			Comment: fromDomainComment(comment),
		},
	}, nil
}

// Delete a comment for an article
// (DELETE /articles/{slug}/comments/{id})
func (s *StrictAPIServer) DeleteArticleComment(
	ctx context.Context,
	request DeleteArticleCommentRequestObject,
) (DeleteArticleCommentResponseObject, error) {
	if err := s.svc.DeleteComment(ctx, request.Slug, request.Id); err != nil {
		return DeleteArticleComment422JSONResponse{}, fmt.Errorf("delete article comment: %w", err)
	}

	return DeleteArticleComment200Response{}, nil
}

// Unfavorite an article
// (DELETE /articles/{slug}/favorite)
func (s *StrictAPIServer) DeleteArticleFavorite(
	ctx context.Context,
	request DeleteArticleFavoriteRequestObject,
) (DeleteArticleFavoriteResponseObject, error) {
	art, err := s.svc.UnfavoriteArticle(ctx, getUserIDFromContext(ctx), request.Slug)
	if err != nil {
		return DeleteArticleFavorite422JSONResponse{}, fmt.Errorf(
			"delete article favorite: %w",
			err,
		)
	}

	return DeleteArticleFavorite200JSONResponse{
		SingleArticleResponseJSONResponse: SingleArticleResponseJSONResponse{
			Article: fromDomainArticle(art),
		},
	}, nil
}

// Favorite an article
// (POST /articles/{slug}/favorite)
func (s *StrictAPIServer) CreateArticleFavorite(
	ctx context.Context,
	request CreateArticleFavoriteRequestObject,
) (CreateArticleFavoriteResponseObject, error) {
	art, err := s.svc.FavoriteArticle(ctx, getUserIDFromContext(ctx), request.Slug)
	if err != nil {
		return CreateArticleFavorite422JSONResponse{}, fmt.Errorf(
			"create article favorite: %w",
			err,
		)
	}

	return CreateArticleFavorite200JSONResponse{
		SingleArticleResponseJSONResponse: SingleArticleResponseJSONResponse{
			Article: fromDomainArticle(art),
		},
	}, nil
}

// Get a profile
// (GET /profiles/{username})
func (s *StrictAPIServer) GetProfileByUsername(
	ctx context.Context,
	request GetProfileByUsernameRequestObject,
) (GetProfileByUsernameResponseObject, error) {
	profile, err := s.svc.GetProfile(ctx, getUserIDFromContext(ctx), request.Username)
	if err != nil {
		return GetProfileByUsername422JSONResponse{}, fmt.Errorf("get profile by username: %w", err)
	}

	return GetProfileByUsername200JSONResponse{
		ProfileResponseJSONResponse: ProfileResponseJSONResponse{
			Profile: FromDomainProfile(profile),
		},
	}, nil
}

// Unfollow a user
// (DELETE /profiles/{username}/follow)
func (s *StrictAPIServer) UnfollowUserByUsername(
	ctx context.Context,
	request UnfollowUserByUsernameRequestObject,
) (UnfollowUserByUsernameResponseObject, error) {
	prof, err := s.svc.UnfollowUser(ctx, getUserIDFromContext(ctx), request.Username)
	if err != nil {
		return UnfollowUserByUsername422JSONResponse{}, fmt.Errorf("unfollow by username: %w", err)
	}

	return UnfollowUserByUsername200JSONResponse{
		ProfileResponseJSONResponse: ProfileResponseJSONResponse{
			Profile: FromDomainProfile(prof),
		},
	}, nil
}

// Follow a user
// (POST /profiles/{username}/follow)
func (s *StrictAPIServer) FollowUserByUsername(
	ctx context.Context,
	request FollowUserByUsernameRequestObject,
) (FollowUserByUsernameResponseObject, error) {
	prof, err := s.svc.FollowUser(ctx, getUserIDFromContext(ctx), request.Username)
	if err != nil {
		return FollowUserByUsername422JSONResponse{}, fmt.Errorf("follow by username: %w", err)
	}

	return FollowUserByUsername200JSONResponse{
		ProfileResponseJSONResponse: ProfileResponseJSONResponse{
			Profile: FromDomainProfile(prof),
		},
	}, nil
}

// Get tags
// (GET /tags)
func (s *StrictAPIServer) GetTags(
	ctx context.Context,
	_ GetTagsRequestObject,
) (GetTagsResponseObject, error) {
	tags, err := s.svc.GetTags(ctx)
	if err != nil {
		return GetTags422JSONResponse{}, fmt.Errorf("get tags: %w", err)
	}

	return GetTags200JSONResponse{
		TagsResponseJSONResponse: TagsResponseJSONResponse{
			Tags: fromDomainTags(tags),
		},
	}, nil
}

// Get current user
// (GET /user)
func (s *StrictAPIServer) GetCurrentUser(
	ctx context.Context,
	_ GetCurrentUserRequestObject,
) (GetCurrentUserResponseObject, error) {
	user, err := s.svc.GetCurrentUser(ctx, getUserIDFromContext(ctx))
	if err != nil {
		return GetCurrentUser422JSONResponse{}, fmt.Errorf("get current user: %w", err)
	}

	return GetCurrentUser200JSONResponse{
		UserResponseJSONResponse: UserResponseJSONResponse{
			User: fromDomainUser(user, getTokenFromContext(ctx)),
		},
	}, nil
}

// Update current user
// (PUT /user)
func (s *StrictAPIServer) UpdateCurrentUser(
	ctx context.Context,
	request UpdateCurrentUserRequestObject,
) (UpdateCurrentUserResponseObject, error) {
	usr, err := s.svc.UpdateUser(
		ctx,
		getUserIDFromContext(ctx),
		request.Body.User.Username,
		request.Body.User.Email,
		request.Body.User.Password,
		request.Body.User.Bio,
		request.Body.User.Image,
	)
	if err != nil {
		return UpdateCurrentUser422JSONResponse{}, fmt.Errorf("update current user: %w", err)
	}

	return UpdateCurrentUser200JSONResponse{
		UserResponseJSONResponse: UserResponseJSONResponse{
			User: fromDomainUser(usr, getTokenFromContext(ctx)),
		},
	}, nil
}

// (POST /users)
func (s *StrictAPIServer) CreateUser(
	ctx context.Context,
	request CreateUserRequestObject,
) (CreateUserResponseObject, error) {
	usr, err := s.svc.RegisterUser(
		ctx,
		uuid.Must(uuid.NewV7()),
		request.Body.User.Username,
		request.Body.User.Email,
		request.Body.User.Password,
	)
	if err != nil {
		return CreateUser422JSONResponse{}, fmt.Errorf("create user: %w", err)
	}

	return CreateUser201JSONResponse{
		UserResponseJSONResponse: UserResponseJSONResponse{
			User: fromDomainUser(usr, ""),
		},
	}, nil
}

// Existing user login
// (POST /users/login)
func (s *StrictAPIServer) Login(
	ctx context.Context,
	request LoginRequestObject,
) (LoginResponseObject, error) {
	usr, _, err := s.svc.AuthUser(ctx, request.Body.User.Email, request.Body.User.Password)
	if err != nil {
		return Login401Response{}, fmt.Errorf("login: %w", err)
	}

	// set a jwt
	_, jws, errE := s.tokenAuth.Encode(map[string]any{
		jwt.SubjectKey:    usr.ID.String(),
		jwt.IssuedAtKey:   time.Now().Unix(),
		jwt.ExpirationKey: time.Now().Add(1 * time.Hour).Unix(),
	})
	if errE != nil {
		return Login422JSONResponse{}, fmt.Errorf("login encode token: %w", errE)
	}

	return &Login200JSONResponse{
		UserResponseJSONResponse: UserResponseJSONResponse{
			User: fromDomainUser(usr, jws),
		},
	}, nil
}
