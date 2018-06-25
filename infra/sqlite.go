package infra

import (
	"database/sql"
	"os"
	"path"

	_ "github.com/mattn/go-sqlite3"
)

var dbMap map[string]*sql.DB

func init() {
	dbMap = map[string]*sql.DB{}
}

func NewSqliteDB(dbName string) (*sql.DB, error) {
	if _, e := dbMap[dbName]; e {
		return dbMap[dbName], nil
	}

	dbPath := os.Getenv("DATABASE_PATH")
	dbFilePath := path.Join(dbPath, dbName+".db")
	db, err := sql.Open("sqlite3", dbFilePath)
	if err != nil {
		return nil, err
	}

	dbMap[dbName] = db
	return db, nil
}
