package cmd

type Env string

const (
	EnvLocal   Env = "local"
	EnvStaging Env = "staging"
	EnvQA      Env = "qa"
	EnvSandbox Env = "sandbox"
	EnvProd    Env = "prod"
)

func ParseEnvFromString(env string) Env {
	switch env {
	case "local":
		return EnvLocal
	case "staging":
		return EnvStaging
	case "qa":
		return EnvQA
	case "sandbox":
		return EnvSandbox
	case "prod":
		return EnvProd
	default:
		return EnvLocal
	}
}

func (e Env) String() string {
	return string(e)
}
