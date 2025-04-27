package db

import (
	"testing"

	"github.com/google/uuid"

	"realworld/internal/domain"
)

func TestRepository_CreateArticle(t *testing.T) {
	t.Parallel()

	testrep := withRepo(t, "create_article")
	t.Cleanup(func() {
		for _, f := range testrep.GetShutdownFuncs() {
			if err := f(t.Context()); err != nil {
				t.Errorf("could not shutdown: %v", err)
			}
		}
	})

	type args struct {
		username    string
		title       string
		description string
		body        string
		tagList     []string
	}

	tests := []struct {
		name    string
		args    args
		want    *domain.Article
		wantErr bool
	}{
		{
			name: "create article with existing user and non existing tags",
			args: args{
				username:    "jakeq",
				title:       "How to train your dragon",
				description: "Ever wonder how?",
				body:        "It takes a Jacobian",
				tagList:     []string{"dragons", "training"},
			},
			want: &domain.Article{
				Slug:        "how-to-train-your-dragon",
				Title:       "How to train your dragon",
				Description: "Ever wonder how?",
				Body:        "It takes a Jacobian",
				TagList:     []domain.Tag{domain.Tag("dragons"), domain.Tag("training")},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			// save a user
			userID, errU := uuid.NewV7()
			if errU != nil {
				t.Errorf("could not generate uuid: %v", errU)
			}

			usr, errUsr := testrep.RegisterUser(t.Context(), userID, tt.args.username, tt.args.username+"@gmail.com", "")
			if errUsr != nil {
				t.Errorf("could not register user: %v", errUsr)
			}

			got, err := testrep.CreateArticle(t.Context(), usr.ID, tt.args.title, tt.args.description, tt.args.body, tt.args.tagList)
			if (err != nil) != tt.wantErr {
				t.Errorf("Repository.CreateArticle() error = %v, wantErr %v", err, tt.wantErr)

				return
			}

			if got == nil ||
				got.Slug != tt.want.Slug ||
				got.Title != tt.want.Title ||
				got.Description != tt.want.Description ||
				got.Body != tt.want.Body {
				t.Errorf("Repository.CreateArticle() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestRepository_GetArticle(t *testing.T) {
	t.Parallel()

	testrep := withRepo(t, "get_article")
	t.Cleanup(func() {
		for _, f := range testrep.GetShutdownFuncs() {
			if err := f(t.Context()); err != nil {
				t.Errorf("could not shutdown: %v", err)
			}
		}
	})

	tests := []struct {
		name    string
		slug    string
		want    *domain.Article
		wantErr bool
	}{
		{
			name: "get article",
			slug: "how-to-train-your-dragon-2",
			want: &domain.Article{
				Slug:           "how-to-train-your-dragon-2",
				Title:          "How to train your dragon 2",
				Description:    "Ever wonder how?",
				Body:           "It takes a Jacobian",
				TagList:        []domain.Tag{domain.Tag("dragons"), domain.Tag("training")},
				Favorited:      false,
				FavoritesCount: 0,
				Author: domain.Profile{
					Username:  "author",
					Bio:       "",
					Following: false,
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			// save a user
			userID, errU := uuid.NewV7()
			if errU != nil {
				t.Errorf("could not generate uuid: %v", errU)
			}

			usr, errUsr := testrep.RegisterUser(t.Context(), userID, tt.want.Author.Username, tt.want.Author.Username+"@gmail.com", "")
			if errUsr != nil {
				t.Errorf("could not register user: %v", errUsr)
			}

			tags := make([]string, len(tt.want.TagList))

			for idx, tag := range tt.want.TagList {
				tags[idx] = string(tag)
			}

			_, _ = testrep.CreateArticle(t.Context(), usr.ID, tt.want.Title, tt.want.Description, tt.want.Body, tags)

			got, err := testrep.GetArticle(t.Context(), usr.ID, tt.slug)
			if (err != nil) != tt.wantErr {
				t.Errorf("Repository.GetArticle() error = %v, wantErr %v", err, tt.wantErr)

				return
			}

			if got == nil ||
				got.Slug != tt.want.Slug ||
				got.Title != tt.want.Title ||
				got.Description != tt.want.Description ||
				got.Body != tt.want.Body ||
				got.Favorited != tt.want.Favorited ||
				got.FavoritesCount != tt.want.FavoritesCount ||
				len(got.TagList) != len(tt.want.TagList) ||
				got.Author.Username != tt.want.Author.Username ||
				got.Author.Bio != tt.want.Author.Bio {
				t.Errorf("Repository.GetArticle() = %v, want %v", got, tt.want)
			}
		})
	}
}
