package test

import (
	"os"
	"testing"

	"github.com/fireworq/fireworq/config"
	"github.com/fireworq/fireworq/test/mysql"
)

func runWithMySQL(block func()) error {
	dsn := config.Get("mysql_dsn")
	config.Set("repository_mysql_dsn", dsn)
	config.Set("queue_mysql_dsn", dsn)

	return mysqltest.With(dsn, block)
}

// Run runs a TestMain for a single "driver" configuration value.
func Run(m *testing.M) (int, error) {
	var status int
	var err error

	if config.Get("driver") == "mysql" {
		err = runWithMySQL(func() {
			status = m.Run()
		})
	} else {
		status = m.Run()
	}

	return status, err
}

// RunAll runs a TestMain for all "driver" configuration values.
func RunAll(m *testing.M) {
	drivers := []string{"mysql", "in-memory"}
	for _, driver := range drivers {
		config.Locally("driver", driver, func() {
			status, err := Run(m)
			if err != nil {
				panic(err)
			}
			if status != 0 {
				os.Exit(status)
			}
		})
	}
}

// If returns if the configuration value of a key matches with one of
// the specified values.
func If(key string, values ...string) bool {
	for _, v := range values {
		if config.Get(key) == v {
			return true
		}
	}
	return false
}
