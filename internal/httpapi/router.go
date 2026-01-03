package httpapi

import (
	"context"
	"log/slog"
	"time"

	"github.com/getkin/kin-openapi/openapi3filter"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/cors"
	"github.com/go-chi/jwtauth/v5"
	"github.com/go-chi/render"
	"github.com/induzo/gocom/http/middleware/writablecontext"
	oapimiddleware "github.com/oapi-codegen/nethttp-middleware"

	"realworld/internal/domain"
)

func CreateRouter(
	ctx context.Context,
	svc domain.APIService,
	logger *slog.Logger,
	isDebug bool,
	jwtSecret string,
) (*chi.Mux, error) {
	// create chi router
	rtr := chi.NewRouter()

	rtr.Use(middleware.Recoverer)
	rtr.Use(middleware.URLFormat)

	if isDebug {
		rtr.Use(
			cors.Handler(
				cors.Options{},
			),
		)
	}

	// normal handlers/operations, with json and timeout
	rtr.Group(func(rtr chi.Router) {
		const endpointTimeout = 60 * time.Second

		// Stop processing after 60 seconds.
		rtr.Use(middleware.Timeout(endpointTimeout))
		// force usage of json in request and response
		rtr.Use(render.SetContentType(render.ContentTypeJSON))

		rtr.Use(writablecontext.Middleware)

		swagger, errSw := GetSwagger()
		if errSw != nil {
			logger.ErrorContext(
				ctx,
				"adminapi oapi router, error loading swagger spec",
				slog.Any("err", errSw),
			)

			return
		}

		swagger.Servers = nil

		oapiRouter := chi.NewRouter()

		jwtA := jwtauth.New("HS256", []byte(jwtSecret), nil)

		// Create an instance of our handler which satisfies the generated interface
		oapiServerStrictHandler := NewStrictHandler(NewStrictAPIServer(svc, jwtA), nil)

		rtr.Use(
			oapimiddleware.OapiRequestValidatorWithOptions(
				swagger,
				&oapimiddleware.Options{
					Options: openapi3filter.Options{
						AuthenticationFunc: NewAuthenticator(jwtA),
					},
				},
			),
		)

		HandlerFromMux(oapiServerStrictHandler, oapiRouter)

		rtr.Mount("/", oapiRouter)
	})

	return rtr, nil
}
