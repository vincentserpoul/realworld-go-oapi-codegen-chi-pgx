package oapi

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/go-chi/jwtauth/v5"
	"github.com/gofrs/uuid/v5"
	"github.com/lestrrat-go/jwx/jwt"

	"realworld/internal/domain"
	"realworld/internal/oapi/oapigen"
)

// StrictServerInterface represents all server handlers.
type Server struct {
	svc       domain.APIService
	tokenAuth *jwtauth.JWTAuth
}

func NewServer(svc domain.APIService, jwtSecret string) *Server {
	tokenAuth := jwtauth.New("HS256", []byte(jwtSecret), nil)

	return &Server{
		svc,
		tokenAuth,
	}
}

// Get recent articles globally
// (GET /articles)
func (s *Server) GetArticles(
	ctx context.Context,
	request oapigen.GetArticlesRequestObject,
) (oapigen.GetArticlesResponseObject, error) {
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
		return oapigen.GetArticles422JSONResponse{}, fmt.Errorf("get articles: %w", errA)
	}

	return oapigen.GetArticles200JSONResponse{
		MultipleArticlesResponseJSONResponse: oapigen.MultipleArticlesResponseJSONResponse{
			Articles:      fromDomainArticles(articles),
			ArticlesCount: len(articles),
		},
	}, nil
}

// Create an article
// (POST /articles)
func (s *Server) CreateArticle(
	ctx context.Context,
	request oapigen.CreateArticleRequestObject,
) (oapigen.CreateArticleResponseObject, error) {
	art, err := s.svc.CreateArticle(
		ctx,
		getUserIDFromContext(ctx),
		request.Body.Article.Title,
		request.Body.Article.Description,
		request.Body.Article.Body,
		*request.Body.Article.TagList,
	)
	if err != nil {
		return oapigen.CreateArticle422JSONResponse{}, fmt.Errorf("create article: %w", err)
	}

	return oapigen.CreateArticle201JSONResponse{
		SingleArticleResponseJSONResponse: oapigen.SingleArticleResponseJSONResponse{
			Article: fromDomainArticle(art),
		},
	}, nil
}

// Get recent articles from users you follow
// (GET /articles/feed)
func (s *Server) GetArticlesFeed(
	ctx context.Context,
	request oapigen.GetArticlesFeedRequestObject,
) (oapigen.GetArticlesFeedResponseObject, error) {
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
		return oapigen.GetArticlesFeed422JSONResponse{}, fmt.Errorf("get articles feed: %w", errA)
	}

	return oapigen.GetArticlesFeed200JSONResponse{
		MultipleArticlesResponseJSONResponse: oapigen.MultipleArticlesResponseJSONResponse{
			Articles:      fromDomainArticles(articles),
			ArticlesCount: len(articles),
		},
	}, nil
}

// Delete an article
// (DELETE /articles/{slug})
func (s *Server) DeleteArticle(
	ctx context.Context,
	request oapigen.DeleteArticleRequestObject,
) (oapigen.DeleteArticleResponseObject, error) {
	if err := s.svc.DeleteArticle(ctx, getUserIDFromContext(ctx), request.Slug); err != nil {
		return oapigen.DeleteArticle422JSONResponse{}, fmt.Errorf("delete article: %w", err)
	}

	return oapigen.DeleteArticle200Response{}, nil
}

// Get an article
// (GET /articles/{slug})
func (s *Server) GetArticle(
	ctx context.Context,
	request oapigen.GetArticleRequestObject,
) (oapigen.GetArticleResponseObject, error) {
	article, errA := s.svc.GetArticle(ctx, getUserIDFromContext(ctx), request.Slug)
	if errA != nil {
		return oapigen.GetArticle422JSONResponse{}, fmt.Errorf("get article: %w", errA)
	}

	return oapigen.GetArticle200JSONResponse{
		SingleArticleResponseJSONResponse: oapigen.SingleArticleResponseJSONResponse{
			Article: fromDomainArticle(article),
		},
	}, nil
}

// Update an article
// (PUT /articles/{slug})
func (s *Server) UpdateArticle(
	ctx context.Context,
	request oapigen.UpdateArticleRequestObject,
) (oapigen.UpdateArticleResponseObject, error) {
	art, err := s.svc.UpdateArticle(
		ctx,
		getUserIDFromContext(ctx),
		request.Slug,
		request.Body.Article.Title,
		request.Body.Article.Description,
		request.Body.Article.Body,
	)
	if err != nil {
		return oapigen.UpdateArticle422JSONResponse{}, fmt.Errorf("update article: %w", err)
	}

	return oapigen.UpdateArticle200JSONResponse{
		SingleArticleResponseJSONResponse: oapigen.SingleArticleResponseJSONResponse{
			Article: fromDomainArticle(art),
		},
	}, nil
}

// Get comments for an article
// (GET /articles/{slug}/comments)
func (s *Server) GetArticleComments(
	ctx context.Context,
	request oapigen.GetArticleCommentsRequestObject,
) (oapigen.GetArticleCommentsResponseObject, error) {
	comments, err := s.svc.GetComments(ctx, getUserIDFromContext(ctx), request.Slug)
	if err != nil {
		return oapigen.GetArticleComments422JSONResponse{}, fmt.Errorf("get article comment: %w", err)
	}

	return oapigen.GetArticleComments200JSONResponse{
		MultipleCommentsResponseJSONResponse: oapigen.MultipleCommentsResponseJSONResponse{
			Comments: fromDomainComments(comments),
		},
	}, nil
}

// Create a comment for an article
// (POST /articles/{slug}/comments)
func (s *Server) CreateArticleComment(
	ctx context.Context,
	request oapigen.CreateArticleCommentRequestObject,
) (oapigen.CreateArticleCommentResponseObject, error) {
	comment, err := s.svc.AddComment(
		ctx,
		getUserIDFromContext(ctx),
		request.Slug,
		request.Body.Comment.Body,
	)
	if err != nil {
		return oapigen.CreateArticleComment422JSONResponse{}, fmt.Errorf("create article comment: %w", err)
	}

	return oapigen.CreateArticleComment200JSONResponse{
		SingleCommentResponseJSONResponse: oapigen.SingleCommentResponseJSONResponse{
			Comment: fromDomainComment(comment),
		},
	}, nil
}

// Delete a comment for an article
// (DELETE /articles/{slug}/comments/{id})
func (s *Server) DeleteArticleComment(
	ctx context.Context,
	request oapigen.DeleteArticleCommentRequestObject,
) (oapigen.DeleteArticleCommentResponseObject, error) {
	if err := s.svc.DeleteComment(ctx, request.Slug, request.Id); err != nil {
		return oapigen.DeleteArticleComment422JSONResponse{}, fmt.Errorf("delete article comment: %w", err)
	}

	return oapigen.DeleteArticleComment200Response{}, nil
}

// Unfavorite an article
// (DELETE /articles/{slug}/favorite)
func (s *Server) DeleteArticleFavorite(
	ctx context.Context,
	request oapigen.DeleteArticleFavoriteRequestObject,
) (oapigen.DeleteArticleFavoriteResponseObject, error) {
	art, err := s.svc.UnfavoriteArticle(ctx, getUserIDFromContext(ctx), request.Slug)
	if err != nil {
		return oapigen.DeleteArticleFavorite422JSONResponse{}, fmt.Errorf("delete article favorite: %w", err)
	}

	return oapigen.DeleteArticleFavorite200JSONResponse{
		SingleArticleResponseJSONResponse: oapigen.SingleArticleResponseJSONResponse{
			Article: fromDomainArticle(art),
		},
	}, nil
}

// Favorite an article
// (POST /articles/{slug}/favorite)
func (s *Server) CreateArticleFavorite(
	ctx context.Context,
	request oapigen.CreateArticleFavoriteRequestObject,
) (oapigen.CreateArticleFavoriteResponseObject, error) {
	art, err := s.svc.FavoriteArticle(ctx, getUserIDFromContext(ctx), request.Slug)
	if err != nil {
		return oapigen.CreateArticleFavorite422JSONResponse{}, fmt.Errorf("create article favorite: %w", err)
	}

	return oapigen.CreateArticleFavorite200JSONResponse{
		SingleArticleResponseJSONResponse: oapigen.SingleArticleResponseJSONResponse{
			Article: fromDomainArticle(art),
		},
	}, nil
}

// Get a profile
// (GET /profiles/{username})
func (s *Server) GetProfileByUsername(
	ctx context.Context,
	request oapigen.GetProfileByUsernameRequestObject,
) (oapigen.GetProfileByUsernameResponseObject, error) {
	profile, err := s.svc.GetProfile(ctx, getUserIDFromContext(ctx), request.Username)
	if err != nil {
		return oapigen.GetProfileByUsername422JSONResponse{}, fmt.Errorf("get profile by username: %w", err)
	}

	return oapigen.GetProfileByUsername200JSONResponse{
		ProfileResponseJSONResponse: oapigen.ProfileResponseJSONResponse{
			Profile: FromDomainProfile(profile),
		},
	}, nil
}

// Unfollow a user
// (DELETE /profiles/{username}/follow)
func (s *Server) UnfollowUserByUsername(
	ctx context.Context,
	request oapigen.UnfollowUserByUsernameRequestObject,
) (oapigen.UnfollowUserByUsernameResponseObject, error) {
	prof, err := s.svc.UnfollowUser(ctx, getUserIDFromContext(ctx), request.Username)
	if err != nil {
		return oapigen.UnfollowUserByUsername422JSONResponse{}, fmt.Errorf("unfollow by username: %w", err)
	}

	return oapigen.UnfollowUserByUsername200JSONResponse{
		ProfileResponseJSONResponse: oapigen.ProfileResponseJSONResponse{
			Profile: FromDomainProfile(prof),
		},
	}, nil
}

// Follow a user
// (POST /profiles/{username}/follow)
func (s *Server) FollowUserByUsername(
	ctx context.Context,
	request oapigen.FollowUserByUsernameRequestObject,
) (oapigen.FollowUserByUsernameResponseObject, error) {
	prof, err := s.svc.FollowUser(ctx, getUserIDFromContext(ctx), request.Username)
	if err != nil {
		return oapigen.FollowUserByUsername422JSONResponse{}, fmt.Errorf("follow by username: %w", err)
	}

	return oapigen.FollowUserByUsername200JSONResponse{
		ProfileResponseJSONResponse: oapigen.ProfileResponseJSONResponse{
			Profile: FromDomainProfile(prof),
		},
	}, nil
}

// Get tags
// (GET /tags)
func (s *Server) GetTags(
	ctx context.Context,
	_ oapigen.GetTagsRequestObject,
) (oapigen.GetTagsResponseObject, error) {
	tags, err := s.svc.GetTags(ctx)
	if err != nil {
		return oapigen.GetTags422JSONResponse{}, fmt.Errorf("get tags: %w", err)
	}

	return oapigen.GetTags200JSONResponse{
		TagsResponseJSONResponse: oapigen.TagsResponseJSONResponse{
			Tags: fromDomainTags(tags),
		},
	}, nil
}

// Get current user
// (GET /user)
func (s *Server) GetCurrentUser(
	ctx context.Context,
	_ oapigen.GetCurrentUserRequestObject,
) (oapigen.GetCurrentUserResponseObject, error) {
	user, err := s.svc.GetCurrentUser(ctx, getUserIDFromContext(ctx))
	if err != nil {
		return oapigen.GetCurrentUser422JSONResponse{}, fmt.Errorf("get current user: %w", err)
	}

	return oapigen.GetCurrentUser200JSONResponse{
		UserResponseJSONResponse: oapigen.UserResponseJSONResponse{
			User: fromDomainUser(user, getTokenFromContext(ctx)),
		},
	}, nil
}

// Update current user
// (PUT /user)
func (s *Server) UpdateCurrentUser(
	ctx context.Context,
	request oapigen.UpdateCurrentUserRequestObject,
) (oapigen.UpdateCurrentUserResponseObject, error) {
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
		return oapigen.UpdateCurrentUser422JSONResponse{}, fmt.Errorf("update current user: %w", err)
	}

	return oapigen.UpdateCurrentUser200JSONResponse{
		UserResponseJSONResponse: oapigen.UserResponseJSONResponse{
			User: fromDomainUser(usr, getTokenFromContext(ctx)),
		},
	}, nil
}

// (POST /users)
func (s *Server) CreateUser(
	ctx context.Context,
	request oapigen.CreateUserRequestObject,
) (oapigen.CreateUserResponseObject, error) {
	usr, err := s.svc.RegisterUser(
		ctx,
		uuid.Must(uuid.NewV7()),
		request.Body.User.Username,
		request.Body.User.Email,
		request.Body.User.Password,
	)
	if err != nil {
		return oapigen.CreateUser422JSONResponse{}, fmt.Errorf("create user: %w", err)
	}

	return oapigen.CreateUser201JSONResponse{
		UserResponseJSONResponse: oapigen.UserResponseJSONResponse{
			User: fromDomainUser(usr, ""),
		},
	}, nil
}

// Existing user login
// (POST /users/login)
func (s *Server) Login(
	ctx context.Context,
	request oapigen.LoginRequestObject,
) (oapigen.LoginResponseObject, error) {
	usr, _, err := s.svc.AuthUser(ctx, request.Body.User.Email, request.Body.User.Password)
	if err != nil {
		return oapigen.Login401Response{}, fmt.Errorf("login: %w", err)
	}

	// set a jwt
	_, jws, errE := s.tokenAuth.Encode(map[string]interface{}{
		jwt.SubjectKey:    usr.ID.String(),
		jwt.IssuedAtKey:   time.Now().Unix(),
		jwt.ExpirationKey: time.Now().Add(1 * time.Hour).Unix(),
	})
	if errE != nil {
		return oapigen.Login422JSONResponse{}, fmt.Errorf("login encode token: %w", errE)
	}

	return &Login200JSONResponse{
		jws: jws,
		UserResponseJSONResponse: oapigen.UserResponseJSONResponse{
			User: fromDomainUser(usr, jws),
		},
	}, nil
}

type Login200JSONResponse struct {
	oapigen.UserResponseJSONResponse
	jws string
}

func (response *Login200JSONResponse) VisitLoginResponse(resp http.ResponseWriter) error {
	resp.Header().Set("Content-Type", "application/json")
	resp.Header().Set("Authorization", "Token "+response.jws)

	resp.WriteHeader(http.StatusOK)

	if err := json.NewEncoder(resp).Encode(response); err != nil {
		return fmt.Errorf("encode login response: %w", err)
	}

	return nil
}
