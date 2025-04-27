package oapi

import (
	"context"
	"fmt"
	"net/http"
	"strings"

	"github.com/getkin/kin-openapi/openapi3filter"
	"github.com/go-chi/jwtauth/v5"
	"github.com/google/uuid"
	"github.com/induzo/gocom/http/middleware/writablecontext"
	"github.com/lestrrat-go/jwx/v2/jwt"
)

func NewAuthenticator(auth *jwtauth.JWTAuth) openapi3filter.AuthenticationFunc {
	return func(ctx context.Context, input *openapi3filter.AuthenticationInput) error {
		return Authenticate(ctx, auth, input)
	}
}

func getTokenFromContext(ctx context.Context) string {
	tokenAny, ok := writablecontext.FromContext(ctx).Get(TokenContextKey)
	if !ok {
		return ""
	}

	if token, ok := tokenAny.(string); ok {
		return token
	}

	return ""
}

func getUserIDFromContext(ctx context.Context) uuid.UUID {
	userIDAny, isValid := writablecontext.FromContext(ctx).Get(UserIDContextKey)
	if !isValid {
		return uuid.Nil
	}

	uids, isValid := userIDAny.(string)
	if !isValid {
		return uuid.Nil
	}

	userID, err := uuid.Parse(uids)
	if err != nil {
		return uuid.Nil
	}

	return userID
}

const (
	UserIDContextKey = "userID"
	TokenContextKey  = "token"
)

type ErrWrongSecSchemeError struct {
	ExpectedScheme string
	FoundScheme    string
}

func (e ErrWrongSecSchemeError) Error() string {
	return fmt.Sprintf("wrong security scheme: expected %s, found %s", e.ExpectedScheme, e.FoundScheme)
}

// Authenticate uses the specified validator to ensure a JWT is valid, then makes
// sure that the claims provided by the JWT match the scopes as required in the API.
func Authenticate(
	_ context.Context,
	auth *jwtauth.JWTAuth,
	input *openapi3filter.AuthenticationInput,
) error {
	// // Our security scheme is named Token, ensure this is the case
	if input.SecuritySchemeName != "Token" {
		return ErrWrongSecSchemeError{ExpectedScheme: "Token", FoundScheme: input.SecuritySchemeName}
	}

	// verifier the JWS, which will also validate the signature.
	token, err := jwtauth.VerifyRequest(
		auth,
		input.RequestValidationInput.Request,
		findTokenInHeader,
	)
	if err != nil {
		return fmt.Errorf("validating JWS: %w", err)
	}

	if err := CheckTokenClaims(input.Scopes, token); err != nil {
		return fmt.Errorf("token claims don't match: %w", err)
	}

	//nolint:contextcheck // context contains a writable context
	// Add the user ID to the context
	reqstore := writablecontext.FromContext(input.RequestValidationInput.Request.Context())
	reqstore.Set(UserIDContextKey, token.Subject())
	reqstore.Set(TokenContextKey, findTokenInHeader(input.RequestValidationInput.Request))

	return nil
}

func CheckTokenClaims(expectedClaims []string, t jwt.Token) error {
	existingClaims := t.PrivateClaims()

	for _, e := range expectedClaims {
		if _, ok := existingClaims[e]; !ok {
			return jwt.ErrMissingRequiredClaim("missing claim: " + e)
		}
	}

	return nil
}

func findTokenInHeader(r *http.Request) string {
	// Get token from authorization header.
	bearer := r.Header.Get("Authorization")

	if len(bearer) > 6 && strings.ToUpper(bearer[0:5]) == "TOKEN" {
		return bearer[6:]
	}

	return ""
}
