package main

import (
	"encoding/json"
	"flag"
	"log"
	"net"
	"net/http"
	"os"
	"os/signal"

	//	"github.com/docker/docker/pkg/plugins"
	"github.com/gorilla/mux"
)

func main() {
	log.Print("starting up")
	var sock string
	flag.StringVar(&sock, "socket", "/run/docker/plugins/alchemy.sock", "")
	flag.Parse()

	l, lErr := net.Listen("unix", sock)
	if lErr != nil {
		log.Fatalf("can't listen on the unix socket %v", lErr)
	}
	go func() {
		signalChan := make(chan os.Signal, 1)
		signal.Notify(signalChan, os.Interrupt)
		<-signalChan
		log.Print("Received an interrupt, stopping services...")
		l.Close()
		os.Remove(sock)
	}()

	r := mux.NewRouter()
	g := r.Methods("GET").Subrouter()
	g.HandleFunc("/", dead)
	http.Handle("/", g)
	// All requests are HTTP POST requests.
	p := r.Methods("POST").Subrouter()
	p.HandleFunc("/Plugin.Activate", activate)
	log.Print("serving")
	http.Serve(l, r)
	log.Print("done serving")
}

func dead(w http.ResponseWriter, r *http.Request) {
	log.Print("dead call")
	http.Error(w, "", http.StatusMethodNotAllowed)
}

type activateResponse struct {
	Implements []string
}

func activate(w http.ResponseWriter, r *http.Request) {
	log.Print("activate call")
	// empty request to /Plugin.Activate
	// {
	//   "Implements": ["HTTPInterceptor"]
	// }

	err := json.NewEncoder(w).Encode(&activateResponse{
		[]string{"HTTPInterceptor"},
	})
	if err != nil {
		log.Fatal("can't make json to connect to docker")
	}
}

/*

{
  "Name": "plugin-example",
  "Addr": "https://example.com/docker/plugin",
  "TLSConfig": {
    "InsecureSkipVerify": false,
    "CAFile": "/usr/shared/docker/certs/example-ca.pem",
    "CertFile": "/usr/shared/docker/certs/example-cert.pem",
    "KeyFile": "/usr/shared/docker/certs/example-key.pem",
  }
}
*/

// The API is versioned via an Accept header, which currently is always set to application/vnd.docker.plugins.v1+json.
