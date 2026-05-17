package domain

import "testing"

func TestGetSlugFromTitle(t *testing.T) {
	t.Parallel()

	type args struct {
		title string
	}

	tests := []struct {
		name string
		args args
		want string
	}{
		{
			name: "get slug from title",
			args: args{
				title: "How to train your dragon",
			},
			want: "how-to-train-your-dragon",
		},
		{
			name: "get slug from title",
			args: args{
				title: "test",
			},
			want: "test",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			if got := GetSlugFromTitle(tt.args.title); got != tt.want {
				t.Errorf("GetSlugFromTitle() = %v, want %v", got, tt.want)
			}
		})
	}
}
