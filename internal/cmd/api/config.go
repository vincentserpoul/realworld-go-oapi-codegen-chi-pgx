package api

import (
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/knadh/koanf/parsers/toml"
	"github.com/knadh/koanf/providers/env"
	"github.com/knadh/koanf/providers/file"
	"github.com/knadh/koanf/v2"
)

type Config struct {
	Env  string `koanf:"env"`
	Name string `koanf:"name"`

	SecretsPath string `koanf:"secrets_path"`

	Log struct {
		IsPretty bool   `koanf:"is_pretty"`
		Level    string `koanf:"level"`
	} `koanf:"log"`

	WithDebugProfiler bool   `koanf:"with_debug_profiler"`
	Version           string `koanf:"version"`

	Observability struct {
		Collector struct {
			Host               string   `koanf:"host"`
			Port               int      `koanf:"port"`
			Headers            []Header `koanf:"headers"`
			IsInsecure         bool     `koanf:"is_insecure"`
			WithMetricsEnabled bool     `koanf:"with_metrics_enabled"`
		} `koanf:"collector"`
	} `koanf:"observability"`

	HTTP HTTPConfig `koanf:"http"`

	DatabaseURL string `koanf:"database_url"`

	Security SecurityConfig `koanf:"security"`
}

type HTTPConfig struct {
	Port     int `koanf:"port"`
	Timeouts struct {
		ReadTimeout       time.Duration `koanf:"read_timeout"`
		ReadHeaderTimeout time.Duration `koanf:"read_header_timeout"`
		WriteTimeout      time.Duration `koanf:"write_timeout"`
		IdleTimeout       time.Duration `koanf:"idle_timeout"`
	}
	HealthEndpoint string `koanf:"health_endpoint"`
}

type Header struct {
	Key   string `koanf:"key"`
	Value string `koanf:"value"`
}

type SecurityConfig struct {
	JWTSecret string `koanf:"jwt_secret"`
}

func ParseConfig(environment string) (*Config, error) {
	konf := koanf.New(".")

	if err := konf.Load(file.Provider("config/api/base.toml"), toml.Parser()); err != nil {
		return nil, fmt.Errorf("failed to load base config: %w", err)
	}

	if err := konf.Load(file.Provider(fmt.Sprintf("config/api/%s.toml", environment)), toml.Parser()); err != nil {
		return nil, fmt.Errorf("failed to load env config from toml: %w", err)
	}

	// check if a config/api/%s.secrets.toml exists and load it if it is
	if val, ok := konf.Get("secrets_path").(string); ok {
		if _, err := os.Stat(val); err == nil {
			if err := konf.Load(
				file.Provider(
					fmt.Sprintf("config/api/%s.secrets.toml", environment),
				),
				toml.Parser(),
			); err != nil {
				return nil, fmt.Errorf("failed to load env config from toml: %w", err)
			}
		}
	}

	if err := konf.Load(env.Provider("", ".", func(s string) string {
		return strings.ReplaceAll(strings.ToLower(strings.TrimPrefix(s, "")), "__", ".")
	}), nil); err != nil {
		return nil, fmt.Errorf("failed to load env config from ENV: %w", err)
	}

	cfg := &Config{}
	if err := konf.Unmarshal("", &cfg); err != nil {
		return nil, fmt.Errorf("failed to unmarshal config: %w", err)
	}

	return cfg, nil
}
