package main

import (
	"database/sql"
	"log"

	_ "github.com/go-sql-driver/mysql"
)

type SessionTable struct {
	db *sql.DB
}

func (sess *SessionTable) Connect(url string) (err error) {

	sess.db, err = sql.Open("mysql", url)
	if err != nil {
		log.Fatal(err)
		return err
	}
	err = sess.db.Ping()
	if err != nil {
		log.Fatal(err)
		return err
	}
	return nil
}

func (sess *SessionTable) QuerySessionAndRetUid(session string) (int, error) {

	query := ` select uid from sessionTable where session = ?`

	row := sess.db.QueryRow(query, session)
	var uidTmp int
	if err := row.Scan(&uidTmp); err != nil {
		return -1, err
	}

	return uidTmp, nil
}
