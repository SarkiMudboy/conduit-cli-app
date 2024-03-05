package main

import (
	"github.com/jmoiron/sqlx"

	_ "github.com/lib/pq"
)


type Dependency struct {
	Db *sqlx.DB
}

func init() {
	
	Db, err := sqlx.Open("postgres", "user=postgres dbname=conduit password=1234 sslmode=disable")
	if err != nil {
		panic(err)
	}
	err = Db.Ping()
	if err != nil {
		panic(err)
	}

	dependency = &Dependency{
		Db: Db,
	}
}

// restructure the DB (for all)
/* 
	drive_member table 
	 - user id
	 - drive id

	DriveMember struct {}
	|
	v
	Drive 
	 - Members *Users -> drive_member.user
*/


