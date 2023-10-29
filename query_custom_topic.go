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
	Subjt          *SubjectTable
	Session        *SessionTable
	queryClient    QueryClient
	SubjectService *SubjectServiceTable
}

type PullCustomTopicData struct {
	Session  string `json:"session"`
	QueryStr string `json:"query"`
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
			log.Print(err)
			w.WriteHeader(http.StatusBadRequest)
			return
		}
		err = json.Unmarshal(data, &pullData)
		if err != nil {
			log.Print(err)
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
		log.Printf("UID [%d] Query %s \n ", uid, pullData.QueryStr)
		//
		yes, err := pull.SubjectService.UidIsExpires(uid)
		if err != nil {
			log.Print("UidIsExpires error  \n")
			w.WriteHeader(http.StatusBadRequest)
			return
		}
		if yes {
			//reject
			log.Printf("Uid %d IsExpired   \n", uid)
			w.WriteHeader(http.StatusBadRequest)
			return

		}
		//get uid , query .
		w.Header().Set("Content-Type", "application/json")

		var topicResult []*Topic

		query_data, err := RequestToQueryServerCustomTopic(pull.queryClient, pullData.QueryStr)
		if err != nil {
			log.Print(err)
			w.WriteHeader(http.StatusBadRequest)
			return
		}
		err = json.Unmarshal(query_data, &topicResult)
		if err != nil {
			log.Print(err)
			w.WriteHeader(http.StatusBadRequest)
			return
		}

		r, err := json.MarshalIndent(topicResult, " ", " ")
		if err != nil {
			log.Print(err)
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
