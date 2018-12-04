package main

import (
	"fmt"
	"html"
	"log"
	"net/http"
	"time"

	"github.com/sevlyar/go-daemon"
)

// To terminate the daemon use:
//  kill `cat pid`

func main() {
	cntxt := &daemon.Context{
		PidFileName: "ingress.pid",
		PidFilePerm: 0644,
		LogFileName: "ingress.log",
		LogFilePerm: 0640,
		WorkDir:     "./",
		Umask:       027,
		Args:        []string{"[go-daemon sample]"},
	}

	d, err := cntxt.Reborn()
	if err != nil {
		log.Println("Unable to run ", err)
		log.Fatal("Already running?")
	}
	if d != nil {
		return
	}
	defer cntxt.Release()

	log.Print("- - - - - - - - - - - - - - -")
	log.Print("daemon started")

	for {
		time.Sleep(1 * time.Second)
	}
}

func serveHTTP() {
	http.HandleFunc("/", httpHandler)
	http.ListenAndServe("127.0.0.1:8080", nil)
}

func httpHandler(w http.ResponseWriter, r *http.Request) {
	log.Printf("request from %s: %s %q", r.RemoteAddr, r.Method, r.URL)
	fmt.Fprintf(w, "go-daemon: %q", html.EscapeString(r.URL.Path))
}
