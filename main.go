package main

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
)

/*

	1) user client-> pull  {session}
	2) server push -> {session(id) , all topic to user }

*/

var queryServe string

type PullNewTopic struct {
	Subjt   *SubjectTable
	Session *SessionTable
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

			return nil, err
		}
		topics = append(topics, tmp)
	}
	return topics, nil

}

func RequestToQueryServer(w http.ResponseWriter, topic string) error {
	//timeStr := strings.Split(time.Now().String(), " ")[0]
	timeStr := "2023-09-06"
	reqStr := fmt.Sprintf(queryServe+"/query?topic=%s&date=%s", topic, timeStr)
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
	w.Header().Set("Content-Type", "application/json")
	io.Copy(w, respon.Body)

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
		//send topicinfo to user.
		for _, topic := range topics {
			if err := RequestToQueryServer(w, topic); err != nil {
				w.WriteHeader(http.StatusBadRequest)
				return
			}
		}
		return
	}

	w.WriteHeader(http.StatusBadRequest)
}

func main() {
	queryServe = os.Getenv("queryurl")
	if queryServe == "" {
		log.Fatal("queryurl is none")
		return
	}

	sqlurl := os.Getenv("sqlurl")
	if sqlurl == "" {
		log.Fatal("sqlurl is none")
		return
	}
	serverPort := os.Getenv("serverport")
	if serverPort == "" {
		log.Fatal("serverPort is none")
		return
	}

	pull := new(PullNewTopic)
	pull.Subjt = new(SubjectTable)
	pull.Session = new(SessionTable)

	if err := pull.Subjt.Connect(sqlurl); err != nil {
		log.Fatal(err)
	}

	if err := pull.Session.Connect(sqlurl); err != nil {
		log.Fatal(err)
	}
	log.Print("connect !")

	mux := http.NewServeMux()
	mux.Handle("/pull_topic", pull)
	server := &http.Server{
		Addr:    serverPort,
		Handler: mux,
	}
	if err := server.ListenAndServe(); err != nil {
		log.Fatal(err)
	}

}
