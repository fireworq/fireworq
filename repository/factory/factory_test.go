package factory

import (
	"testing"

	"github.com/fireworq/fireworq/config"
)

func TestInvalidDriver(t *testing.T) {
	config.Locally("driver", "nothing", func() {
		defer func() {
			if r := recover(); r == nil {
				t.Error("It should die")
			}
		}()

		NewRepositories()
	})
}

func TestInvalidDsn(t *testing.T) {
	config.Locally("driver", "mysql", func() {
		config.Locally("repository_mysql_dsn", "xxxx", func() {
			defer func() {
				if r := recover(); r == nil {
					t.Error("It should die")
				}
			}()

			NewRepositories()
		})

		config.Locally("repository_mysql_dsn", "", func() {
			config.Locally("mysql_dsn", "xxxx", func() {
				defer func() {
					if r := recover(); r == nil {
						t.Error("It should die")
					}
				}()

				NewRepositories()
			})
		})
	})
}
