package main

import (
	"log"
	"net/http"
	"os"
)

/*

	1) user client-> pull  {session}
	2) server push -> {session(id) , all topic to user }

*/

var queryServe string

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
	redisurl := os.Getenv("redisurl")
	if redisurl == "" {
		log.Fatal("redisurl is none")
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
	log.Print("pull connect !")

	pullCustom := new(PullCustomTopic)
	pullCustom.Subjt = new(SubjectTable)
	pullCustom.Session = new(SessionTable)
	pullCustom.RedisW = new(RedisWrapper)

	if err := pullCustom.Subjt.Connect(sqlurl); err != nil {
		log.Fatal(err)
	}

	if err := pullCustom.Session.Connect(sqlurl); err != nil {
		log.Fatal(err)
	}
	log.Print("pullCustom connect !")

	if err := pullCustom.RedisW.Connect(redisurl); err != nil {
		log.Fatal(err)
	}
	log.Print("redis connect !")

	mux := http.NewServeMux()
	mux.Handle("/pull_topic", pull)
	mux.Handle("/pull_topic_custom", pullCustom)

	server := &http.Server{
		Addr:    serverPort,
		Handler: mux,
	}
	if err := server.ListenAndServe(); err != nil {
		log.Fatal(err)
	}

}
