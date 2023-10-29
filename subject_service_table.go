package main

import (
	"database/sql"
	"log"
	"time"
)

type SubjectServiceTable struct {
	db *sql.DB
}

func (st *SubjectServiceTable) Connect(url string) (err error) {

	st.db, err = sql.Open("mysql", url)
	if err != nil {
		log.Fatal(err)
		return err
	}
	err = st.db.Ping()
	if err != nil {
		log.Fatal(err)
		return err
	}
	return nil
}

func (st *SubjectServiceTable) UidIsExpires(uid int) (bool, error) {
	row := st.db.QueryRow("select ifnull (max(expires_date),\"NO SUBJECT\") from subjectTable where uid=?", uid)

	var date string
	err := row.Scan(&date)
	if err != nil {
		return false, err
	}
	if date == "NO SUBJECT" {
		return true, nil
	}
	if date < time.Now().UTC().String() {
		return true, nil
	}
	return false, nil
}
