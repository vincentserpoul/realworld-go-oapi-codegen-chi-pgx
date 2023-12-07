package api

import "time"

type Config struct {
	Env  string `koanf:"env"`
	Name string `koanf:"name"`

	Log struct {
		IsPretty bool   `koanf:"isPretty"`
		Level    string `koanf:"level"`
	} `koanf:"log"`

	WithDebugProfiler bool   `koanf:"withDebugProfiler"`
	Version           string `koanf:"version"`

	Observability struct {
		Collector struct {
			Host          string `koanf:"host"`
			Port          int    `koanf:"port"`
			IsSecure      bool   `koanf:"isSecure"`
			EnableMetrics bool   `koanf:"enableMetrics"`
		} `koanf:"collector"`
	} `koanf:"observability"`

	HTTP HTTPConfig `koanf:"http"`

	DatabaseURL string `koanf:"databaseURL"`

	Security SecurityConfig `koanf:"security"`
}

type HTTPConfig struct {
	Port     int `koanf:"port"`
	Timeouts struct {
		ReadTimeout       time.Duration `koanf:"readTimeout"`
		ReadHeaderTimeout time.Duration `koanf:"readHeaderTimeout"`
		WriteTimeout      time.Duration `koanf:"writeTimeout"`
		IdleTimeout       time.Duration `koanf:"idleTimeout"`
	}
	HealthEndpoint string `koanf:"healthEndpoint"`
}

type SecurityConfig struct {
	JWTSecret string `koanf:"jwtSecret"`
}
