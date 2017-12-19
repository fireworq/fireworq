package mysql

import (
	"os"
	"testing"
	"time"

	"github.com/fireworq/fireworq/config"
	"github.com/fireworq/fireworq/jobqueue"
	"github.com/fireworq/fireworq/model"
	"github.com/fireworq/fireworq/test"
	"github.com/fireworq/fireworq/test/jobqueue"
	"github.com/fireworq/fireworq/test/mysql"
)

func TestMain(m *testing.M) {
	config.Locally("driver", "mysql", func() {
		status, err := test.Run(m)
		if err != nil {
			panic(err)
		}
		os.Exit(status)
	})
}

// Common tests

func TestNew(t *testing.T) {
	_ = New(&model.Queue{Name: "test", MaxWorkers: 30}, "dummy")
}

func TestSubtests(t *testing.T) {
	jqtest.TestSubtests(t, runSubtests)
}

// MySQL specific tests

func TestNode(t *testing.T) {
	jq := New(&model.Queue{Name: "test", MaxWorkers: 30}, Dsn())
	jq.Start()
	defer func() { <-jq.Stop() }()

	hasNodeInfo, ok := jq.(jobqueue.HasNodeInfo)
	if !ok {
		t.Error("Must have Node() method")
	}
	node, err := hasNodeInfo.Node()
	if err != nil {
		t.Error(err)
	}
	if len(node.ID) <= 0 {
		t.Error("Must return an ID")
	}
	if len(node.Host) <= 0 {
		t.Error("Must return a host name")
	}
}

func runSubtests(t *testing.T, db, q string, tests []jqtest.Subtest) {
	dsn := Dsn()

	jq := New(&model.Queue{Name: q, MaxWorkers: 30}, dsn)
	jq.Start()
	defer func() { <-jq.Stop() }()
	time.Sleep(500 * time.Millisecond) // wait for up

	for _, test := range tests {
		err := mysqltest.TruncateTables(dsn)
		if err != nil {
			t.Error(err)
		}
		test(t, jq)
	}
}
