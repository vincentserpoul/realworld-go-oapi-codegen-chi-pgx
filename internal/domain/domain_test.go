package domain

import (
	"flag"
	"os"
	"testing"

	"go.uber.org/goleak"
)

func TestMain(m *testing.M) {
	leak := flag.Bool("leak", false, "use leak detector")

	flag.Parse()

	if *leak {
		goleak.VerifyTestMain(
			m,
			goleak.IgnoreAnyFunction(
				"github.com/testcontainers/testcontainers-go.(*Reaper).Connect.func1",
			),
		)

		return
	}

	os.Exit(m.Run())
}
