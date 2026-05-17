package cmd

import (
	"fmt"
	"os"
	"strings"

	"github.com/knadh/koanf/parsers/toml"
	"github.com/knadh/koanf/providers/env"
	"github.com/knadh/koanf/providers/file"
	"github.com/knadh/koanf/v2"
)

type BasicConfigurator interface {
	GetBasicConfig() BasicConfig
}

type BasicConfig struct {
	Env     string `koanf:"env"`
	Name    string `koanf:"name"`
	Version string `koanf:"version"`

	SecretsPath string `koanf:"secrets_path"`

	Log struct {
		IsPretty bool   `koanf:"is_pretty"`
		Level    string `koanf:"level"`
	} `koanf:"log"`

	WithDebugProfiler bool `koanf:"with_debug_profiler"`

	Observability observabilityConfig

	HTTP httpConfig `koanf:"http"`
}

// ParseConfig parse a config file into a Config struct, in a generic manner
// that can be used by any binary.
// The binaryName is the name of the binary, and the environment is the
// environment to load the config for.
// The config is loaded in the following order:
// 1. config/%s/base.toml
// 2. config/%s/%s.toml
// 3. ENV
// 4. secrets_path (if it exists)
// The config is unmarshalled into the Config struct.
// The config is returned along with an error if any.
//

func ParseConfig[Config BasicConfigurator](binaryName, environment Env) (Config, error) {
	var emptyConf Config

	konf := koanf.New(".")

	if err := konf.Load(file.Provider(fmt.Sprintf("config/%s/base.toml", binaryName)), toml.Parser()); err != nil {
		return emptyConf, fmt.Errorf(
			"failed to load base config (%s, %s): %w",
			binaryName,
			environment,
			err,
		)
	}

	if err := konf.Load(
		file.Provider(fmt.Sprintf("config/%s/%s.toml", binaryName, environment)),
		toml.Parser(),
	); err != nil {
		return emptyConf, fmt.Errorf(
			"failed to load env config from toml (%s, %s): %w",
			binaryName,
			environment,
			err,
		)
	}

	if err := konf.Load(env.Provider("", ".", func(s string) string {
		return strings.ReplaceAll(strings.ToLower(strings.TrimPrefix(s, "")), "__", ".")
	}), nil); err != nil {
		return emptyConf, fmt.Errorf(
			"failed to load env config from ENV (%s, %s): %w",
			binaryName,
			environment,
			err,
		)
	}

	// check if a config/worker/%s.secrets.toml exists and load it if it is
	if val, ok := konf.Get("secrets_path").(string); ok {
		if _, err := os.Stat(val); err == nil {
			if err := konf.Load(
				file.Provider(val),
				toml.Parser(),
			); err != nil {
				return emptyConf,
					fmt.Errorf(
						"failed to load secrets config from toml(%s, %s) in file `%s`: %w",
						binaryName, environment, val, err)
			}
		}
	}

	var cfg Config
	if err := konf.Unmarshal("", &cfg); err != nil {
		return emptyConf, fmt.Errorf(
			"failed to unmarshal config(%s, %s): %w",
			binaryName,
			environment,
			err,
		)
	}

	bCfg := cfg.GetBasicConfig()

	if bCfg.Name == "" || bCfg.Version == "" || bCfg.Env == "" {
		return emptyConf,
			MissingRequiredFieldsError{
				Name:    bCfg.Name,
				Version: bCfg.Version,
				Env:     bCfg.Env,
			}
	}

	return cfg, nil
}

type MissingRequiredFieldsError struct {
	BinaryName string
	Env        string
	Name       string
	Version    string
}

func (m MissingRequiredFieldsError) Error() string {
	return fmt.Sprintf(
		"missing required fields in config(%s, %s): name `%s`, version `%s`, env `%s`",
		m.BinaryName, m.Env, m.Name, m.Version, m.Env,
	)
}
