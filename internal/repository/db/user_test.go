package db

import (
	"context"
	"strconv"
	"testing"

	"github.com/gofrs/uuid/v5"

	"realworld/internal/domain"
)

func TestRepository_RegisterUser(t *testing.T) {
	t.Parallel()

	testrep := withRepo(t, "register_user")
	t.Cleanup(func() {
		for _, f := range testrep.GetShutdownFuncs() {
			if err := f(context.Background()); err != nil {
				t.Errorf("could not shutdown: %v", err)
			}
		}
	})

	type args struct {
		username string
		email    string
		password string
	}

	tests := []struct {
		name    string
		args    args
		want    *domain.User
		wantErr bool
	}{
		{
			name: "register user",
			args: args{
				username: "jakeregister",
				email:    "jake@po.com",
				password: "123456",
			},
			want: &domain.User{
				Username: "jakeregister",
				Email:    "jake@po.com",
				Password: "123456",
			},
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			userID, _ := uuid.NewV7()

			got, err := testrep.RegisterUser(context.Background(), userID, tt.args.username, tt.args.email, tt.args.password)
			if (err != nil) != tt.wantErr {
				t.Errorf("Repository.RegisterUser() error = %v, wantErr %v", err, tt.wantErr)

				return
			}

			if got.Username != tt.want.Username || got.Email != tt.want.Email || got.Password != tt.want.Password {
				t.Errorf("Repository.RegisterUser() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestRepository_UpdateUser(t *testing.T) {
	t.Parallel()

	testrep := withRepo(t, "update_user")
	t.Cleanup(func() {
		for _, f := range testrep.GetShutdownFuncs() {
			if err := f(context.Background()); err != nil {
				t.Errorf("could not shutdown: %v", err)
			}
		}
	})

	type args struct {
		username *string
		password *string
		email    *string
		bio      *string
		image    *string
	}

	baseUser := &domain.User{
		Username: "jakeupdate",
		Password: "123456",
		Email:    "jakeupdate@gmail.com",
		Image:    "https://avatars.githubusercontent.com/u/32737308?v=4",
		Bio:      "",
	}

	tests := []struct {
		name    string
		args    args
		want    *domain.User
		wantErr bool
	}{
		{
			name: "update user username",
			args: args{
				username: func() *string { s := "updatepopol"; return &s }(),
			},
			want: &domain.User{
				Username: "updatepopol",
				Password: baseUser.Password,
				Email:    "0" + baseUser.Email,
				Bio:      baseUser.Bio,
				Image:    baseUser.Image,
			},
		},
		{
			name: "update user email",
			args: args{
				email: func() *string { s := "updatedemails@po.com"; return &s }(),
			},
			want: &domain.User{
				Username: "1" + baseUser.Username,
				Password: baseUser.Password,
				Email:    "updatedemails@po.com",
				Bio:      baseUser.Bio,
				Image:    baseUser.Image,
			},
		},
		{
			name: "update user image",
			args: args{
				image: func() *string { s := "https://static.productionready.io/images/smiley-cyrus.jpg"; return &s }(),
			},
			want: &domain.User{
				Username: "2" + baseUser.Username,
				Password: baseUser.Password,
				Email:    "2" + baseUser.Email,
				Bio:      baseUser.Bio,
				Image:    "https://static.productionready.io/images/smiley-cyrus.jpg",
			},
		},
		{
			name: "update user bio",
			args: args{
				bio: func() *string { s := "I work at statefarm 2"; return &s }(),
			},
			want: &domain.User{
				Username: "3" + baseUser.Username,
				Password: baseUser.Password,
				Email:    "3" + baseUser.Email,
				Bio:      "I work at statefarm 2",
				Image:    baseUser.Image,
			},
		},
	}

	for i, tt := range tests {
		tt := tt
		i := i

		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			// create another empty user and copy the base user to it
			currUser := &domain.User{}
			*currUser = *baseUser

			currUser.ID = uuid.Must(uuid.NewV7())
			currUser.Username = strconv.Itoa(i) + currUser.Username
			currUser.Email = strconv.Itoa(i) + currUser.Email

			// register user
			currU, errRU := testrep.RegisterUser(
				context.Background(),
				currUser.ID,
				currUser.Username,
				currUser.Email,
				currUser.Password,
			)
			if errRU != nil {
				t.Errorf("could not register user: %v", errRU)

				return
			}

			got, err := testrep.UpdateUser(context.Background(), currU.ID, tt.args.username, tt.args.email, tt.args.password, tt.args.bio, tt.args.image)
			if (err != nil) != tt.wantErr {
				t.Errorf("Repository.UpdateUser() error = %v, wantErr %v", err, tt.wantErr)

				return
			}

			if got.Username != tt.want.Username || got.Email != tt.want.Email || got.Password != tt.want.Password || got.Bio != tt.want.Bio || got.Image != tt.want.Image {
				t.Errorf("Repository.UpdateUser() = %v, want %v", got, tt.want)
			}
		})
	}
}
