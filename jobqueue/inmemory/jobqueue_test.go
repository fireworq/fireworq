package inmemory

import (
	"os"
	"testing"

	"github.com/fireworq/fireworq/config"
	"github.com/fireworq/fireworq/test"
	"github.com/fireworq/fireworq/test/jobqueue"
)

func TestMain(m *testing.M) {
	config.Locally("driver", "in-memory", func() {
		status, err := test.Run(m)
		if err != nil {
			panic(err)
		}
		os.Exit(status)
	})
}

// Common tests

func TestNew(t *testing.T) {
	_ = New()
}

func TestSubtests(t *testing.T) {
	jqtest.TestSubtests(t, runSubtests)
}

// in-memory specific tests

func runSubtests(t *testing.T, db, q string, tests []jqtest.Subtest) {
	for _, test := range tests {
		jq := New()
		jq.Start()
		test(t, jq)
		jq.Stop()
	}
}
