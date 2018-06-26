package infra

import (
	"database/sql"
	"path"

	_ "github.com/mattn/go-sqlite3"
	"github.com/yuki-eto/5ch-slack-bot/config"
)

var dbMap map[string]*sql.DB

func init() {
	dbMap = map[string]*sql.DB{}
}

func NewSqliteDB(dbName string) (*sql.DB, error) {
	if _, e := dbMap[dbName]; e {
		return dbMap[dbName], nil
	}

	cfg := config.GetEnvConfig()
	dbFilePath := path.Join(cfg.DatabasePath, dbName + ".db")
	db, err := sql.Open("sqlite3", dbFilePath)
	if err != nil {
		return nil, err
	}

	dbMap[dbName] = db
	return db, nil
}
