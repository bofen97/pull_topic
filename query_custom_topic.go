package main

import (
	context "context"
	"encoding/json"
	"io"
	"log"
	"net/http"
	"net/url"
)

type PullCustomTopic struct {
	Subjt       *SubjectTable
	Session     *SessionTable
	RedisW      *RedisWrapper
	queryClient QueryClient
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

func RequestToQueryServerCustomTopic(queryClient QueryClient, topic string) ([]byte, error) {
	topic = url.QueryEscape(topic)

	log.Printf("Gen Request Topic [%s] And Go GRPC . \n", topic)
	rts, err := queryClient.QueryCustom(context.Background(), &QueryCustomArg{Topic: topic})
	if err != nil {
		log.Print(err)
		return nil, err
	}
	data, err := json.Marshal(rts.GetQuerys())
	if err != nil {
		log.Print(err)
		return nil, err
	}

	return data, nil
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
		var topicResults []*Topic

		for _, topic := range topics {
			var topicResult []*Topic

			// str, err := pull.RedisW.GetKey(topic)
			// if err == nil {
			// log.Print(len(str))
			// w.Write([]byte(str))
			// log.Printf("Cache Hited %s \n", topic)
			// continue
			// }

			data, err := RequestToQueryServerCustomTopic(pull.queryClient, topic)
			if err != nil {
				w.WriteHeader(http.StatusBadRequest)
				return
			}
			err = json.Unmarshal(data, &topicResult)
			if err != nil {
				w.WriteHeader(http.StatusBadRequest)
				return
			}
			topicResults = append(topicResults, topicResult...)
		}
		r, err := json.MarshalIndent(topicResults, " ", " ")
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			return
		}
		w.Write(r)
		return
	}

	w.WriteHeader(http.StatusBadRequest)
}

type Topic struct {
	Title     string `json:"title"`
	Url       string `json:"url"`
	Summary   string `json:"summary"`
	Authors   string `json:"authors"`
	Published string `json:"published"`
}
