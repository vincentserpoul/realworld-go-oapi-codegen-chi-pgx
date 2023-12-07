package writablecontext

import (
	"context"
	"net/http"
)

type Key string

const contextKey Key = "writeablecontext"

type Store map[string]any

func newStore() Store {
	return make(Store)
}

func (s Store) Set(key string, value any) {
	if s == nil {
		s = newStore()
	}

	s[key] = value
}

func (s Store) Get(key string) (any, bool) {
	val, ok := s[key]

	return val, ok
}

func Middleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		newR := r.WithContext(context.WithValue(r.Context(), contextKey, newStore()))
		next.ServeHTTP(w, newR)
	})
}

func FromContext(ctx context.Context) Store {
	currStore, ok := ctx.Value(contextKey).(Store)
	if !ok {
		return nil
	}

	return currStore
}
