package main

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/gorilla/mux"
	"github.com/gorilla/websocket"
)

// Command represents typical command from client to tank
type Command struct {
	Commands string `json:"commands"`
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

	r := mux.NewRouter().StrictSlash(true)

	r.HandleFunc("/api/tanks", p.listTanks).Methods("GET")
	r.HandleFunc("/api/tanks/{id}", p.getTank).Methods("GET")
	r.HandleFunc("/api/tanks/{id}/connect", p.getTankConnection).Methods("GET")

	r.HandleFunc("/api/connect", p.handleConnect).Methods("GET")
	r.HandleFunc("/api/connect/{name}", p.handleConnect).Methods("GET")

	r.PathPrefix("/static/").Handler(http.StripPrefix("/static/", http.FileServer(http.Dir("./static/"))))

	r.HandleFunc("/", redirectToStatic).Methods("GET")

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

// getTank returns single tank by id
func (p *PitankServer) getTank(w http.ResponseWriter, r *http.Request) {
	id := getStringVar(r, "id")
	if id == nil {
		http.Error(w, "id is not passed", http.StatusBadRequest)
		return
	}

	tank, exist := p.Tanks[*id]
	if !exist {
		http.Error(w, "Not Found", http.StatusNotFound)
		return
	}

	enc := json.NewEncoder(w)
	err := enc.Encode(tank)
	if err != nil {
		http.Error(w, "Can't encode message:"+err.Error(), http.StatusInternalServerError)
		return
	}
}

// getTankConnection establishes websocket connection to the tank
func (p *PitankServer) getTankConnection(w http.ResponseWriter, r *http.Request) {
	id := getStringVar(r, "id")
	if id == nil {
		http.Error(w, "id is not passed", http.StatusBadRequest)
		return
	}

	tank, exist := p.Tanks[*id]
	if !exist {
		http.Error(w, "Not Found", http.StatusNotFound)
		return
	}

	// upgrade connection to websocket
	// to use it as bidirectional command channel
	conn, err := p.wsUpgrader.Upgrade(w, r, nil)
	if err != nil {
		fmt.Println(err)
		return
	}

	for {
		_, data, err := conn.ReadMessage()
		if err != nil {
			msg := "Error on read: " + err.Error()
			fmt.Println(msg)
			return
		}

		var c Command
		err = json.Unmarshal(data, &c)
		if err != nil {
			fmt.Println("Error on command unmarshal:", err.Error())
			continue
		}
		fmt.Println("Sending command", c, "to tank", tank.Name)
		tank.SendCommand(c)
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

	pitank, exist := p.Tanks[*name]
	if !exist {
		// generate and register new pitank
		pitank = NewPitank(*name)
		p.Tanks[*name] = pitank
	}

	// upgrade connection to websocket
	// to use it as bidirectional command channel
	conn, err := p.wsUpgrader.Upgrade(w, r, nil)
	if err != nil {
		fmt.Println(err)
		return
	}
	pitank.Connect(conn)

	// send pitank info (for debug)
	pitank.SendCommand(pitank)

	// serve writes from command chan to websocket in gorouting
	go pitank.WritePump()
	go pitank.ReadPump()
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

// redirectToStatic redirects to static html landing page
func redirectToStatic(w http.ResponseWriter, r *http.Request) {
	http.Redirect(w, r, "/static/index.html", http.StatusFound)
}
