package main

import (
	"log"
	"net/http"
	"os"

	serviceRegisterCenter "github.com/bofen97/ServiceRegister"
	grpc "google.golang.org/grpc"
)

/*

	1) user client-> pull  {session}
	2) server push -> {session(id) , all topic to user }

*/

func main() {
	etcdserver := os.Getenv("etcdserver")
	if etcdserver == "" {
		log.Fatal("etcdserver is none")
		return
	}
	serverPort := os.Getenv("serverport")
	if serverPort == "" {
		log.Fatal("serverPort is none")
		return
	}
	sqlurl := os.Getenv("sqlurl")
	if sqlurl == "" {
		log.Fatal("sqlurl is none")
		return
	}

	src, err := serviceRegisterCenter.NewRegisteService([]string{
		etcdserver,
	}, 5)
	if err != nil {
		log.Fatal(err)
	}
	go src.WatchService("query_topic")
	go src.PutServiceAddr("pull_topic", "pull_topic"+serverPort)
	go src.ListenLaser()

	conn, err := grpc.Dial(src.GetK("query_topic"), grpc.WithInsecure())
	if err != nil {
		log.Fatal(err)
	}
	defer conn.Close()
	qclient := NewQueryClient(*conn)

	pull := new(PullNewTopic)
	pull.Subjt = new(SubjectTable)
	pull.Session = new(SessionTable)
	pull.subjectService = new(SubjectServiceTable)
	pull.queryClient = qclient

	if err := pull.subjectService.Connect(sqlurl); err != nil {
		log.Fatal(err)
	}
	if err := pull.Subjt.Connect(sqlurl); err != nil {
		log.Fatal(err)
	}

	if err := pull.Session.Connect(sqlurl); err != nil {
		log.Fatal(err)
	}
	log.Print("pull connect !")

	pullCustom := new(PullCustomTopic)
	pullCustom.SubjectService = new(SubjectServiceTable)
	pullCustom.Subjt = new(SubjectTable)
	pullCustom.Session = new(SessionTable)
	pullCustom.queryClient = qclient

	if err := pullCustom.SubjectService.Connect(sqlurl); err != nil {
		log.Fatal(err)
	}
	if err := pullCustom.Subjt.Connect(sqlurl); err != nil {
		log.Fatal(err)
	}

	if err := pullCustom.Session.Connect(sqlurl); err != nil {
		log.Fatal(err)
	}
	log.Print("pullCustom connect !")

	mux := http.NewServeMux()
	mux.Handle("/pull_topic", pull)
	mux.Handle("/query_custom", pullCustom)

	server := &http.Server{
		Addr:    serverPort,
		Handler: mux,
	}
	if err := server.ListenAndServe(); err != nil {
		log.Fatal(err)
	}

}
