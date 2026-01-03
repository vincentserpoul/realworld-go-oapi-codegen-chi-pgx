package cmd

import "testing"

func TestParseEnvFromString(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		env  string
		want Env
	}{
		{
			name: "local",
			env:  "local",
			want: EnvLocal,
		},
		{
			name: "staging",
			env:  "staging",
			want: EnvStaging,
		},
		{
			name: "qa",
			env:  "qa",
			want: EnvQA,
		},
		{
			name: "sandbox",
			env:  "sandbox",
			want: EnvSandbox,
		},
		{
			name: "prod",
			env:  "prod",
			want: EnvProd,
		},
		{
			name: "default",
			env:  "unknown",
			want: EnvLocal,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			if got := ParseEnvFromString(tt.env); got != tt.want {
				t.Errorf("ParseEnvFromString() = %v, want %v", got, tt.want)
			}
		})
	}
}
