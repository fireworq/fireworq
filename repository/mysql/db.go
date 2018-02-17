//go:generate go-assets-builder -p mysql -o assets.go ../../data/repository

package mysql

import (
	"database/sql"
	"io/ioutil"
	"time"

	"github.com/fireworq/fireworq/config"

	_ "github.com/go-sql-driver/mysql" // initialize the driver
	"github.com/rs/zerolog/log"
)

var schema []string

func init() {
	schema = []string{
		"/data/repository/mysql/schema/queue.sql",
		"/data/repository/mysql/schema/routing.sql",
		"/data/repository/mysql/schema/config_revision.sql",
	}
}

// Dsn returns the data source name of the storage specified in the
// configuration.
func Dsn() string {
	dsn := config.Get("repository_mysql_dsn")
	if dsn != "" {
		return dsn
	}
	return config.Get("mysql_dsn")
}

// NewDB creates an instance of DB handler.
func NewDB() (*sql.DB, error) {
	dsn := Dsn()

	log.Info().Msgf("Connecting database %s ...", dsn)

	db, err := sql.Open("mysql", dsn)
	if err != nil {
		return nil, err
	}
	func() {
		var timeout int
		if db.QueryRow("SELECT @@SESSION.wait_timeout").Scan(&timeout) != nil {
			return
		}

		t := timeout - 1
		if t < 1 {
			t = 1
		}
		log.Debug().Msgf("wait_timeout: %d", timeout)
		db.SetConnMaxLifetime(time.Duration(t) * time.Second)
	}()

	for _, path := range schema {
		f, err := Assets.Open(path)
		if err != nil {
			log.Panic().Msg(err.Error())
		}

		query, err := ioutil.ReadAll(f)
		f.Close()
		if err != nil {
			log.Panic().Msg(err.Error())
		}

		_, err = db.Exec(string(query))
		if err != nil {
			return nil, err
		}
	}

	return db, nil
}
