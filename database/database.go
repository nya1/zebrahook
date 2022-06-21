package database

import (
	"database/sql"

	"github.com/spf13/viper"
)

type DatabaseInfo struct {
	OpenedDatabase *sql.DB
	Dsn            string
	DbType         string
}

func Open() DatabaseInfo {
	// connect to db
	dbType := viper.GetString("database.type")
	dsn := viper.GetString("database.dsn")

	db, err := sql.Open(dbType, dsn)
	if err != nil {
		panic(err)
	}
	return DatabaseInfo{
		OpenedDatabase: db,
		DbType:         dbType,
		Dsn:            dsn,
	}
}
