package main

import (
	"github.com/jmoiron/sqlx"

	_ "github.com/lib/pq"
)

var Db *sqlx.DB

func init() {
	var err error	
	Db, err = sqlx.Open("postgres", "user=postgres dbname=conduit password=1234 sslmode=disable")
	if err != nil {
		panic(err)
	}
	err = Db.Ping()
	if err != nil {
		panic(err)
	}
}



