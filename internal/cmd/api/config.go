package api

import "time"

type Config struct {
	Env  string `koanf:"env"`
	Name string `koanf:"name"`

	Log struct {
		IsPretty bool   `koanf:"is_pretty"`
		Level    string `koanf:"level"`
	} `koanf:"log"`

	WithDebugProfiler bool   `koanf:"with_debug_profiler"`
	Version           string `koanf:"version"`

	Observability struct {
		Collector struct {
			Host          string `koanf:"host"`
			Port          int    `koanf:"port"`
			IsSecure      bool   `koanf:"is_secure"`
			EnableMetrics bool   `koanf:"enable_metrics"`
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

type SecurityConfig struct {
	JWTSecret string `koanf:"jwt_secret"`
}
