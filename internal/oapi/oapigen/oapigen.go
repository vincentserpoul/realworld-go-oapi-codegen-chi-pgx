package oapigen

//go:generate go tool oapi-codegen --config=server.cfg.yml ../../../api/openapi.yml
//go:generate go tool oapi-codegen --config=types.cfg.yml ../../../api/openapi.yml
