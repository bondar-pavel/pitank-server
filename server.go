package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/gorilla/mux"
)

// Pitank structure stores connected tank details
type Pitank struct {
	Name               string
	Status             string
	LastRegistration   time.Time
	LastDeregistration time.Time
	//CommandChan
}

// PitankServer configures webserver
type PitankServer struct {
	Port  string
	Tanks map[string]*Pitank
}

func NewPitankServer(port string) *PitankServer {
	return &PitankServer{
		Port:  port,
		Tanks: make(map[string]*Pitank),
	}
}

// Serve initialize webserver routing
func (p *PitankServer) Serve() {
	r := mux.NewRouter()

	r.HandleFunc("/api/tanks", p.listTanks).Methods("GET")

	fmt.Println("Starting server on port", p.Port)
	err := http.ListenAndServe(":"+p.Port, r)
	if err != nil {
		fmt.Println("Error on starting server:", err)
		return
	}
}

// listTanks returns list of connected tanks
func (p *PitankServer) listTanks(w http.ResponseWriter, r *http.Request) {
	tanks := make([]*Pitank, 0)
	for _, tank := range p.Tanks {
		tanks = append(tanks, tank)
	}

	enc := json.NewEncoder(w)
	err := enc.Encode(tanks)
	if err != nil {
		http.Error(w, "Can't encode message:"+err.Error(), http.StatusInternalServerError)
		return
	}
}
