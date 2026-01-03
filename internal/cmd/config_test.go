package cmd

import (
	"testing"

	"github.com/google/go-cmp/cmp"
)

type ConfigTest struct {
	BasicConfig BasicConfig `koanf:",squash"`
	DatabaseURL string      `koanf:"database_url"`
}

func (cfg ConfigTest) GetBasicConfig() BasicConfig {
	return cfg.BasicConfig
}

func TestParseConfig(t *testing.T) {
	t.Parallel()

	type args struct {
		binaryName  Env
		environment Env
	}

	tests := []struct {
		name    string
		args    args
		want    ConfigTest
		wantErr bool
	}{
		{
			name: "success",
			args: args{
				binaryName:  "test",
				environment: "local",
			},
			want: ConfigTest{
				BasicConfig: BasicConfig{
					Env:               "local",
					Name:              "test",
					Version:           "v0",
					SecretsPath:       "config/adminapi/local.secrets.toml",
					WithDebugProfiler: true,
					Log: struct {
						IsPretty bool   `koanf:"is_pretty"`
						Level    string `koanf:"level"`
					}{
						IsPretty: true,
						Level:    "debug",
					},
					Observability: observabilityConfig{
						Collector: struct {
							Host       string   `koanf:"host"`
							Port       int      `koanf:"port"`
							Headers    []Header `koanf:"headers"`
							IsInsecure bool     `koanf:"is_insecure"`
						}{
							Host: "127.0.0.1",
						},
					},
				},
				DatabaseURL: "127.0.0.1",
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			got, err := ParseConfig[ConfigTest](tt.args.binaryName, tt.args.environment)
			if (err != nil) != tt.wantErr {
				t.Errorf("ParseConfig() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !cmp.Equal(got, tt.want) {
				t.Errorf("ParseConfig() diff = %v", cmp.Diff(got, tt.want))
			}
		})
	}
}

func TestMissingRequiredFieldsError_Error(t *testing.T) {
	t.Parallel()

	err := MissingRequiredFieldsError{
		BinaryName: "test",
		Env:        "local",
		Name:       "name",
		Version:    "1",
	}

	want := "missing required fields in config(test, local): name `name`, version `1`, env `local`"

	if got := err.Error(); got != want {
		t.Errorf("Error() = %v, want %v", got, want)
	}
}
