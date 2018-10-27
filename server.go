package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/gorilla/mux"
	"github.com/gorilla/websocket"
)

// Pitank structure stores connected tank details
type Pitank struct {
	Name               string    `json:"name"`
	Status             string    `json:"status"`
	LastRegistration   time.Time `json:"last_registration"`
	LastDeregistration time.Time `json:"last_deregistration"`
	//CommandChan
}

return &Pitank{
	func NewPitank(name string) *Pitank {
		Name:             name,
		Status:           "connected",
		LastRegistration: time.Now(),
	}
}

// PitankServer configures webserver
type PitankServer struct {
	Port       string
	Tanks      map[string]*Pitank
	wsUpgrader websocket.Upgrader
}

func NewPitankServer(port string) *PitankServer {
	return &PitankServer{
		Port:  port,
		Tanks: make(map[string]*Pitank),
	}
}

// Serve initialize webserver routing
func (p *PitankServer) Serve() {
	p.wsUpgrader = websocket.Upgrader{}

	r := mux.NewRouter()

	r.HandleFunc("/api/tanks", p.listTanks).Methods("GET")
	r.HandleFunc("/api/connect", p.handleConnect).Methods("GET")
	r.HandleFunc("/api/connect/{name}", p.handleConnect).Methods("GET")

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

// handleConnect initialize websocket with pitank
// and registers pitank on the server
func (p *PitankServer) handleConnect(w http.ResponseWriter, r *http.Request) {
	name := getStringVar(r, "name")
	if name == nil {
		msg := "Error: request should contain pitank name"
		fmt.Println(msg)
		http.Error(w, msg, http.StatusBadRequest)
		return
	}

	// generate and register new pitank
	pitank := NewPitank(*name)
	p.Tanks[*name] = pitank

	// upgrate pitank connection to websocket
	// to use it as bidirectional command channel
	conn, err := p.wsUpgrader.Upgrade(w, r, nil)
	if err != nil {
		fmt.Println(err)
		return
	}
	// for now just do nothing and close connection
	defer conn.Close()

	pitank.LastDeregistration = time.Now()
}

// getStringVar return pointer to value of the variable,
// if variable is not found, nil is returned
func getStringVar(r *http.Request, varName string) (value *string) {
	vars := mux.Vars(r)
	if val, ok := vars[varName]; ok {
		value = &val
	}
	return
}
