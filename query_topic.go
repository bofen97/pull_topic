package main

import (
	"context"
	"encoding/json"
	"io"
	"log"
	"net/http"
	"strings"
	"time"
)

type PullNewTopic struct {
	Subjt       *SubjectTable
	Session     *SessionTable
	queryClient QueryClient
}

type PullNewTopicData struct {
	Session string `json:"session"`
}

func (pull *PullNewTopic) GetTopicStr(uid int) ([]string, error) {

	rows, err := pull.Subjt.db.Query("select distinct topic from userSubjectTable where uid=?", uid)
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

func RequestToQueryServer(w http.ResponseWriter, queryClient QueryClient, topic string) error {
	preDay := time.Now().Add(-24 * time.Hour)
	timeStr := strings.Split(preDay.String(), " ")[0]

	rts, err := queryClient.QueryTopic(context.Background(), &QueryTopicArg{
		Topic: topic,
		Date:  timeStr,
	})
	if err != nil {
		log.Print(err)
		return err
	}
	data, err := json.Marshal(rts.GetQuerys())
	if err != nil {
		log.Print(err)
		return err
	}
	w.Write(data)

	return nil
}

func (pull *PullNewTopic) ServeHTTP(w http.ResponseWriter, r *http.Request) {

	if r.Method == "POST" && r.Header.Get("Content-Type") == "application/json" {

		var pullData PullNewTopicData

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
			if err := RequestToQueryServer(w, pull.queryClient, topic); err != nil {
				w.WriteHeader(http.StatusBadRequest)
				return
			}
		}
		return
	}

	w.WriteHeader(http.StatusBadRequest)
}
