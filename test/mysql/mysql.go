package mysqltest

import (
	"database/sql"
	"fmt"

	"github.com/go-sql-driver/mysql"
)

// With runs a block with locking the DB and truncating all tables in
// the DB.
func With(dsn string, block func()) error {
	db, err := sql.Open("mysql", dsn)
	if err != nil {
		return err
	}
	defer db.Close()

	lockName := "fireworq_mysqltest"
	_, err = db.Exec(fmt.Sprintf(`SELECT GET_LOCK('%s', 65536)`, lockName))
	if err != nil {
		return err
	}
	defer func() {
		db.Exec(fmt.Sprintf(`SELECT RELEASE_LOCK('%s')`, lockName))
	}()

	TruncateTables(dsn)

	block()
	return nil
}

// TruncateTables truncates all tables in the DB specified by a DSN.
func TruncateTables(dsn string) error {
	cfg, err := mysql.ParseDSN(dsn)
	if err != nil {
		return err
	}
	dbName := cfg.DBName

	db, err := sql.Open("mysql", dsn)
	if err != nil {
		return err
	}
	defer db.Close()

	rows, err := db.Query(fmt.Sprintf(
		`SELECT TABLE_NAME FROM INFORMATION_SCHEMA.TABLES
         WHERE TABLE_SCHEMA = '%s'`,
		dbName,
	))
	if err != nil {
		return err
	}
	defer rows.Close()

	for rows.Next() {
		var table string
		if err := rows.Scan(&table); err != nil {
			return err
		}
		_, err = db.Exec(fmt.Sprintf("DELETE FROM `%s`", table))
	}
	return err
}
