package models

import (
	"database/sql"
	"log"
)

// OpenDB ..
func OpenDB(source string) *sql.DB {
	var err error
	db, err := sql.Open("postgres", source)
	if err != nil {
		log.Panic(err)
	}

	if err = db.Ping(); err != nil {
		log.Panic(err)
	}

	// this is just here until I move all my old code into the models package
	return db
}
