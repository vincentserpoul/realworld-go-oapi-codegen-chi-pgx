package db

import (
	"context"
	"reflect"
	"testing"

	"github.com/google/uuid"

	"realworld/internal/domain"
)

func TestRepository_GetProfile(t *testing.T) {
	t.Parallel()

	testrep := withRepo(t, "get_profile")
	t.Cleanup(func() {
		for _, f := range testrep.GetShutdownFuncs() {
			if err := f(context.Background()); err != nil {
				t.Errorf("could not shutdown: %v", err)
			}
		}
	})

	tests := []struct {
		name    string
		want    *domain.Profile
		wantErr bool
	}{
		{
			name: "get existing profile",
			want: &domain.Profile{
				Username:  "jakeprofile",
				Bio:       "I work at statefarm",
				Image:     "https://static.productionready.io/images/smiley-cyrus.jpg",
				Following: false,
			},
		},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			// register other user to follow
			regUser, _ := testrep.RegisterUser(context.Background(), uuid.Must(uuid.NewV7()), tt.want.Username, "jakeprofile@lop.com", "122")
			testrep.UpdateUser(context.Background(), regUser.ID, nil, nil, nil, &tt.want.Bio, &tt.want.Image)

			got, err := testrep.GetProfile(context.Background(), regUser.ID, tt.want.Username)
			if (err != nil) != tt.wantErr {
				t.Errorf("Repository.GetProfile() error = %v, wantErr %v", err, tt.wantErr)

				return
			}

			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Repository.GetProfile() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestRepository_FollowUser(t *testing.T) {
	t.Parallel()

	testrep := withRepo(t, "follow_user")
	t.Cleanup(func() {
		for _, f := range testrep.GetShutdownFuncs() {
			if err := f(context.Background()); err != nil {
				t.Errorf("could not shutdown: %v", err)
			}
		}
	})

	follower, _ := testrep.RegisterUser(context.Background(), uuid.Must(uuid.NewV7()), "jakefollow", "jakefollow@po.com", "122")
	followee, _ := testrep.RegisterUser(context.Background(), uuid.Must(uuid.NewV7()), "jakefollowee", "jakefollowee@po.com", "122")

	followeeProfile := &domain.Profile{
		Username:  followee.Username,
		Bio:       followee.Bio,
		Image:     followee.Image,
		Following: true,
	}

	tests := []struct {
		name         string
		existingUser bool
		want         *domain.Profile
		wantErr      bool
	}{
		{
			name:         "follow existing user",
			existingUser: true,
			want:         followeeProfile,
			wantErr:      false,
		},
		{
			name:         "follow non existing user",
			existingUser: false,
			wantErr:      true,
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			followeeUserName := followee.Username
			if !tt.existingUser {
				followeeUserName = "notextist"
			}

			got, err := testrep.FollowUser(context.Background(), follower.ID, followeeUserName)
			if (err != nil) != tt.wantErr {
				t.Errorf("Repository.FollowUser() error = %v, wantErr %v", err, tt.wantErr)

				return
			}

			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Repository.FollowUser() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestRepository_UnfollowUser(t *testing.T) {
	t.Parallel()

	testrep := withRepo(t, "unfollow_user")
	t.Cleanup(func() {
		for _, f := range testrep.GetShutdownFuncs() {
			if err := f(context.Background()); err != nil {
				t.Errorf("could not shutdown: %v", err)
			}
		}
	})

	follower, _ := testrep.RegisterUser(context.Background(), uuid.Must(uuid.NewV7()), "jakefollow", "jakefollow@po.com", "122")
	followee, _ := testrep.RegisterUser(context.Background(), uuid.Must(uuid.NewV7()), "jakefollowee", "jakefollowee@po.com", "122")

	tests := []struct {
		name          string
		wantFollowing bool
		wantErr       bool
	}{
		{
			name:          "unfollow existing user",
			wantFollowing: false,
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			// first follow user
			testrep.FollowUser(context.Background(), follower.ID, followee.Username)

			// then unfollow
			got, err := testrep.UnfollowUser(context.Background(), follower.ID, followee.Username)
			if (err != nil) != tt.wantErr {
				t.Errorf("Repository.UnfollowUser() error = %v, wantErr %v", err, tt.wantErr)

				return
			}

			if got.Following != tt.wantFollowing {
				t.Errorf("Repository.UnfollowUser() = %v, want %v", got, tt.wantFollowing)
			}
		})
	}
}
