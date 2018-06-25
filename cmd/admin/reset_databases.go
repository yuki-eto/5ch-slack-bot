package main

import (
	"fmt"

	_ "github.com/yuki-eto/5ch-slack-bot/dao"
	"github.com/yuki-eto/5ch-slack-bot/infra"
)

func main() {
	dbMaps := infra.GetDbMaps()
	for dbName, dbMap := range dbMaps {
		fmt.Printf("drop tables DB:[%s]\n", dbName)
		if err := dbMap.DropTablesIfExists(); err != nil {
			panic(err)
		}
		fmt.Printf("create tables DB:[%s]\n", dbName)
		if err := dbMap.CreateTablesIfNotExists(); err != nil {
			panic(err)
		}
		if err := dbMap.CreateIndex(); err != nil {
			panic(err)
		}
	}
}
