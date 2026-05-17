package httpapi

//go:generate go tool oapi-codegen --config=./oapi.server.yaml -o ./server.gen.go ./openapi.yml
//go:generate go tool oapi-codegen --config=./oapi.types.yaml -o ./types.gen.go ./openapi.yml
