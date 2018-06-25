package infra

import (
	"fmt"
	"log"

	"gopkg.in/gorp.v2"
)

var dbMaps map[string]*gorp.DbMap

func init() {
	dbMaps = map[string]*gorp.DbMap{}
}

func NewDBMap(dbName string) *gorp.DbMap {
	if dbMap, e := dbMaps[dbName]; e {
		return dbMap
	}

	db, err := NewSqliteDB(dbName)
	if err != nil {
		panic(err)
	}

	dbMap := &gorp.DbMap{Db: db, Dialect: gorp.SqliteDialect{}}
	isDebug := false
	if isDebug {
		dbMap.TraceOn(fmt.Sprintf("[gorp SQL %s]", dbName), &MyGorpLogger{})
	}
	dbMaps[dbName] = dbMap

	return dbMap
}

func GetDbMaps() map[string]*gorp.DbMap {
	return dbMaps
}

type MyGorpLogger struct {
}

func (p *MyGorpLogger) Printf(format string, v ...interface{}) {
	log.Printf(format, v...)
}
