package database

import (
	"errors"
	"testing"

	"github.com/google/go-cmp/cmp"
	"go.uber.org/goleak"
)

func TestMain(m *testing.M) {
	goleak.VerifyTestMain(m)
}

func TestNewSourceInstanceError(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name     string
		err      error
		expected string
	}{
		{
			name:     "nil error",
			err:      nil,
			expected: "could not create new source instance: ",
		},
		{
			name:     "non-nil error",
			err:      errors.New("some error"),
			expected: "could not create new source instance: some error",
		},
		{
			name:     "error with database URL",
			err:      errors.New("some error with postgres://user:password@localhost:5432/dbname"),
			expected: "could not create new source instance: some error with postgres://x:x@localhost:5432/dbname",
		},
		{
			name:     "error with database URL and other text",
			err:      errors.New("some error with postgres://user:password@localhost:5432/dbname and some other text"),
			expected: "could not create new source instance: some error with postgres://x:x@localhost:5432/dbname and some other text",
		},
		{
			name:     "error with database URL pg",
			err:      errors.New("some error with postgres://user:password@localhost:5432/dbname"),
			expected: "could not create new source instance: some error with postgres://x:x@localhost:5432/dbname",
		},
		{
			name:     "error with database URL pg and other text",
			err:      errors.New("some error with postgres://user:password@localhost:5432/dbname and some other text"),
			expected: "could not create new source instance: some error with postgres://x:x@localhost:5432/dbname and some other text",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			err := &NewSourceInstanceError{Err: tc.err}

			if !cmp.Equal(err.Error(), tc.expected) {
				t.Errorf("got: %s, diff: %s", err.Error(), cmp.Diff(err.Error(), tc.expected))
			}
		})
	}
}
