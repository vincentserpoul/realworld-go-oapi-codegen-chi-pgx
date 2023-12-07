package oapi

import (
	"fmt"

	"github.com/getkin/kin-openapi/openapi3filter"
	"github.com/go-chi/chi/v5"
	"github.com/induzo/gocom/http/middleware/writablecontext"
	middleware "github.com/oapi-codegen/nethttp-middleware"

	"realworld/internal/cmd/api"
	"realworld/internal/domain"
	"realworld/internal/oapi/oapigen"
)

func RegisterSvc(
	registerer api.HTTPSvcRegisterer,
	apiSvc domain.APIService,
	jwtSecret string,
) error {
	swagger, err := oapigen.GetSwagger()
	if err != nil {
		return fmt.Errorf("loading swagger spec: %w", err)
	}

	// Clear out the servers array in the swagger spec, that skips validating
	// that server names match. We don't know how this thing will be run.
	swagger.Servers = nil

	// Create an instance of our handler which satisfies the generated interface
	oapiServer := NewServer(apiSvc, jwtSecret)
	oapiServerStrictHandler := oapigen.NewStrictHandler(oapiServer, nil)

	// This is how you set up a basic chi router

	oapiRouter := chi.NewRouter()
	oapiRouter.Use(writablecontext.Middleware)

	// Use our validation middleware to check all requests against the
	// OpenAPI schema.
	oapiRouter.Use(
		middleware.OapiRequestValidatorWithOptions(
			swagger,
			&middleware.Options{
				Options: openapi3filter.Options{
					AuthenticationFunc: NewAuthenticator(oapiServer.tokenAuth),
				},
			},
		),
	)

	// We now register our petStore above as the handler for the interface
	oapigen.HandlerFromMux(oapiServerStrictHandler, oapiRouter)

	if err := registerer.RegisterHTTPSvc(
		"/",
		oapiRouter,
		apiSvc.GetHealthChecks(),
		apiSvc.GetShutdownFuncs(),
	); err != nil {
		return fmt.Errorf("failed to register http svc oapi: %w", err)
	}

	return nil
}
