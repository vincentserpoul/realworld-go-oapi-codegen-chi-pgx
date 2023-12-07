package db

import (
	"context"
	"testing"

	"github.com/gofrs/uuid/v5"

	"realworld/internal/domain"
)

func TestRepository_AddComment(t *testing.T) {
	t.Parallel()

	testrep := withRepo(t, "add_comment")
	t.Cleanup(func() {
		for _, f := range testrep.GetShutdownFuncs() {
			if err := f(context.Background()); err != nil {
				t.Errorf("could not shutdown: %v", err)
			}
		}
	})

	// save a user
	userID, errU := uuid.NewV7()
	if errU != nil {
		t.Errorf("could not generate uuid: %v", errU)
	}

	usr, errUsr := testrep.RegisterUser(context.Background(), userID, "joko", "joko@gmail.com", "")
	if errUsr != nil {
		t.Errorf("could not register user: %v", errUsr)
	}

	// save an article
	art, errA := testrep.CreateArticle(context.Background(), usr.ID, "How to train your dragon", "Ever wonder how?", "It takes a Jacobian", []string{"dragons", "training"})
	if errA != nil {
		t.Errorf("could not create an article: %v", errA)
	}

	type args struct {
		slug string
		body string
	}

	tests := []struct {
		name    string
		args    args
		want    *domain.Comment
		wantErr bool
	}{
		{
			name: "add comment to an article",
			args: args{
				slug: art.Slug,
				body: "I like this article",
			},
			want: &domain.Comment{
				Body: "I like this article",
				Author: domain.Profile{
					Username: usr.Username,
				},
			},
		},
		{
			name: "add comment to a non existing article",
			args: args{
				slug: "non ex",
				body: "I like this article",
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			got, err := testrep.AddComment(context.Background(), usr.ID, tt.args.slug, tt.args.body)
			if (err != nil) != tt.wantErr {
				t.Errorf("Repository.AddComment() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if tt.wantErr {
				return
			}

			if tt.want == nil || tt.want.Body != got.Body || tt.want.Author.Username != got.Author.Username {
				t.Errorf("Repository.AddComment() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestRepository_GetComments(t *testing.T) {
	t.Parallel()

	testrep := withRepo(t, "get_comments")
	t.Cleanup(func() {
		for _, f := range testrep.GetShutdownFuncs() {
			if err := f(context.Background()); err != nil {
				t.Errorf("could not shutdown: %v", err)
			}
		}
	})

	// save a user
	userID, errU := uuid.NewV7()
	if errU != nil {
		t.Errorf("could not generate uuid: %v", errU)
	}

	usr, errUsr := testrep.RegisterUser(context.Background(), userID, "joko", "joko@gmail.com", "")
	if errUsr != nil {
		t.Errorf("could not register user: %v", errUsr)
	}

	// save an article
	art, errA := testrep.CreateArticle(context.Background(), usr.ID, "How to train your dragon", "Ever wonder how?", "It takes a Jacobian", []string{"dragons", "training"})
	if errA != nil {
		t.Errorf("could not create an article: %v", errA)
	}

	tests := []struct {
		name    string
		want    []*domain.Comment
		wantErr bool
	}{
		{
			name: "get comments from an article",
			want: []*domain.Comment{
				{
					Body: "I like this article",
					Author: domain.Profile{
						Username: usr.Username,
					},
				},
			},
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			// create a comment
			_, errC := testrep.AddComment(context.Background(), usr.ID, art.Slug, "I like this article")
			if errC != nil {
				t.Errorf("could not create a comment: %v", errC)

				return
			}

			got, err := testrep.GetComments(context.Background(), usr.ID, art.Slug)
			if (err != nil) != tt.wantErr {
				t.Errorf("Repository.GetComments() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if tt.want == nil ||
				len(tt.want) != len(got) ||
				tt.want[0].Body != got[0].Body ||
				tt.want[0].Author.Username != got[0].Author.Username {
				t.Errorf("Repository.AddComment() = %v, want %v", got, tt.want)
			}
		})
	}
}
