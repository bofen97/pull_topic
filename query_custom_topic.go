package main

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
)

type PullCustomTopic struct {
	Subjt   *SubjectTable
	Session *SessionTable
}

type PullCustomTopicData struct {
	Session string `json:"session"`
}

func (pull *PullCustomTopic) GetTopicStr(uid int) ([]string, error) {

	rows, err := pull.Subjt.db.Query("select distinct customlabel from userSubjectTable where uid=?", uid)
	if err != nil {
		log.Print(err)
		return nil, err
	}
	var topics []string
	for rows.Next() {
		var tmp string
		err = rows.Scan(&tmp)
		if err != nil {
			log.Print(err)

			continue
		}
		topics = append(topics, tmp)
	}
	return topics, nil

}

func RequestToQueryServerCustomTopic(w http.ResponseWriter, topic string) error {
	reqStr := fmt.Sprintf(queryServe+"/query_custom?topic=%s", topic)
	log.Printf("Gen Request Str %s \n", reqStr)

	req, err := http.NewRequest("GET", reqStr, nil)
	if err != nil {
		log.Print(err)
		return err
	}

	respon, err := http.DefaultClient.Do(req)
	if err != nil {
		log.Print(err)
		return err
	}
	io.Copy(w, respon.Body)
	return nil
}

func (pull *PullCustomTopic) ServeHTTP(w http.ResponseWriter, r *http.Request) {

	if r.Method == "POST" && r.Header.Get("Content-Type") == "application/json" {

		var pullData PullCustomTopicData

		data, err := io.ReadAll(r.Body)
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			return
		}
		err = json.Unmarshal(data, &pullData)
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			return
		}
		//check form session
		uid, err := pull.Session.QuerySessionAndRetUid(pullData.Session)
		if err != nil {
			log.Printf("Session [%s] error  \n", pullData.Session)
			w.WriteHeader(http.StatusBadRequest)
			return
		}

		//search subject for uid
		topics, err := pull.GetTopicStr(uid)
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		//send topicinfo to user.
		for _, topic := range topics {
			if err := RequestToQueryServerCustomTopic(w, topic); err != nil {
				w.WriteHeader(http.StatusBadRequest)
				return
			}
		}
		return
	}

	w.WriteHeader(http.StatusBadRequest)
}
