package main

import (
	"encoding/json"
	"github.com/gorilla/mux"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"time"
)

const maxPings = 20
const fileName = "lastpings.json"

type LastPings struct {
	Pings []Ping
}
type Ping struct {
	Stamp time.Time `json:"time"`
	Ip    string    `json:"ip"`
}

func PingHandler(w http.ResponseWriter, r *http.Request) {
	//get the IP
	rAdd := r.Header.Get("X-FORWARDED-FOR")
	if rAdd == "" {
		rAdd = r.RemoteAddr
	}

	//curernt ping
	ping := Ping{time.Now(), rAdd}

	//storage
	tmpDir := os.TempDir()
	tmpFile := filepath.Join(tmpDir, fileName)

	//read last saved
	data, err := ioutil.ReadFile(tmpFile)
	pings := LastPings{}
	if err == nil {
		json.Unmarshal(data, &pings)
	}

	l := len(pings.Pings)
	if l > maxPings {
		//append current ping and remove first
		pings.Pings = append(pings.Pings[l-maxPings:], ping)
	} else {
		//append current ping
		pings.Pings = append(pings.Pings, ping)
	}

	//save
	pingsB, _ := json.Marshal(pings)

	if errw := ioutil.WriteFile(tmpFile, pingsB, 0666); errw != nil {
		log.Fatal(errw)
		w.WriteHeader(http.StatusInternalServerError)
	} else {
		w.WriteHeader(http.StatusOK)

	}
}

func PingsHandler(w http.ResponseWriter, r *http.Request) {
	tmpDir := os.TempDir()
	tmpFile := filepath.Join(tmpDir, fileName)
	dat, err := ioutil.ReadFile(tmpFile)
	if err != nil {
		w.WriteHeader(http.StatusNotFound)
	} else {
		w.Header().Set("Content-Type", "application/json")
		w.Write(dat)
	}
}

func main() {
	r := mux.NewRouter()
	r.HandleFunc("/", PingsHandler)
	r.HandleFunc("/pings", PingsHandler)
	r.HandleFunc("/ping", PingHandler)

	log.Fatal(http.ListenAndServe(":8080", r))
}
