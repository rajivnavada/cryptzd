package crypto

import (
	"database/sql"
	_ "github.com/mattn/go-sqlite3"
	"gopkg.in/gorp.v1"
)

func initDb(filePath string) (*gorp.DbMap, err) {
	db, err := sql.Open("sqlite3", filePath)
	if err != nil {
		return err
	}

	return &gorp.DbMap{Db: db, Dialect: gorp.SqliteDialect{}}
}
