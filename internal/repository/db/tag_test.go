package db

import (
	"reflect"
	"testing"

	"github.com/google/uuid"

	"realworld/internal/domain"
)

func TestRepository_GetTags(t *testing.T) {
	t.Parallel()

	testrep := withRepo(t, "get_tags")
	t.Cleanup(func() {
		for _, f := range testrep.GetShutdownFuncs() {
			if err := f(t.Context()); err != nil {
				t.Errorf("could not shutdown: %v", err)
			}
		}
	})

	tests := []struct {
		name    string
		want    []domain.Tag
		wantErr bool
	}{
		{
			name: "get tags",
			want: []domain.Tag{"tag1", "tag2"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			// insert user
			usr, errU := testrep.RegisterUser(t.Context(), uuid.Must(uuid.NewV7()), "jake", "123@po.com", "123456")
			if errU != nil {
				t.Errorf("Repository.RegisterUser() error = %v", errU)

				return
			}

			// insert articles with tags
			if _, err := testrep.CreateArticle(t.Context(), usr.ID, "title", "description", "body", []string{"tag1", "tag2"}); err != nil {
				t.Errorf("Repository.CreateArticle() error = %v", err)

				return
			}

			got, err := testrep.GetTags(t.Context())
			if (err != nil) != tt.wantErr {
				t.Errorf("Repository.GetTags() error = %v, wantErr %v", err, tt.wantErr)

				return
			}

			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Repository.GetTags() = %v, want %v", got, tt.want)
			}
		})
	}
}
