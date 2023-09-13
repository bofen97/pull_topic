package main

import (
	"database/sql"
	"log"
)

type SubjectTable struct {
	db *sql.DB
}

func (sub *SubjectTable) Connect(url string) (err error) {

	sub.db, err = sql.Open("mysql", url)
	if err != nil {
		log.Fatal(err)
		return err
	}
	err = sub.db.Ping()
	if err != nil {
		log.Fatal(err)
		return err
	}
	return nil
}
