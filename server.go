package main

import (
	"encoding/json"
	"fmt"
	"html/template"
	"io/ioutil"
	"net/http"
	"time"

	"github.com/gorilla/mux"
	"github.com/gorilla/websocket"
)

// Command represents typical command from client to tank
type Command struct {
	Commands string `json:"commands"`
	Offer    string `json:"offer,omitempty"`
	Answer   string `json:"answer,omitempty"`
	Time     int64  `json:"time,omitempty"`
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

	// Public facing api
	r.HandleFunc("/api/tanks", p.listTanks).Methods("GET")
	r.HandleFunc("/api/tanks/{id}", p.getTank).Methods("GET")
	r.HandleFunc("/api/tanks/{id}/offer", p.offerToTank).Methods("POST")
	r.HandleFunc("/api/tanks/{id}/connect", p.clientToTanksWS).Methods("GET")

	// API for connecting pitank to server via WebSocket
	r.HandleFunc("/api/connect", p.handleConnect).Methods("GET")
	r.HandleFunc("/api/connect/{name}", p.handleConnect).Methods("GET")

	r.PathPrefix("/static/").Handler(http.StripPrefix("/static/", http.FileServer(http.Dir("./static/"))))

	r.HandleFunc("/", p.renderTanks).Methods("GET")

	fmt.Println("Starting server on port", p.Port)
	err := http.ListenAndServe(":"+p.Port, r)
	if err != nil {
		fmt.Println("Error on starting server:", err)
		return
	}
}

func (p *PitankServer) renderTanks(w http.ResponseWriter, r *http.Request) {
	tmpl, err := template.ParseFiles("./templates/tanks.html")
	if err != nil {
		http.Error(w, "Can not parse template", http.StatusInternalServerError)
		return
	}

	tanks := make([]*Pitank, 0)
	for _, tank := range p.Tanks {
		tanks = append(tanks, tank)
	}
	tmpl.Execute(w, tanks)
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

func (p *PitankServer) offerToTank(w http.ResponseWriter, r *http.Request) {
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

	offer, err := ioutil.ReadAll(r.Body)
	if err != nil {
		http.Error(w, "Can't read body", http.StatusBadRequest)
		return
	}

	cmd := Command{Offer: string(offer)}
	tank.SendCommand(cmd)
}

// clientToTanksWS establishes websocket connection from the client to the tank (via server)
func (p *PitankServer) clientToTanksWS(w http.ResponseWriter, r *http.Request) {
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
	defer conn.Close()

	go func() {
		ticker := time.NewTicker(time.Second)
		for {
			select {
			case cmd, ok := <-tank.ReplyChan:
				if !ok {
					// Reply channel is closed, terminate connection
					ticker.Stop()
					conn.WriteMessage(websocket.CloseMessage, []byte{})
					return
				}
				err := conn.WriteJSON(cmd)
				if err != nil {
					fmt.Println("Error on ws write, exiting")
					ticker.Stop()
					return
				}
			case t := <-ticker.C:
				ok := false
				tank, ok = p.Tanks[*id]
				if !ok {
					fmt.Println("Tank no longer exits, exiting")
					ticker.Stop()
					return
				}
				var diff time.Duration
				if tank.Status == "connected" {
					diff = t.Sub(tank.LastRegistration)
				} else {
					diff = t.Sub(tank.LastDeregistration)
				}
				msg := fmt.Sprintf("status %s, for %s", tank.Status, diff)
				err := conn.WriteMessage(1, []byte(msg))
				if err != nil {
					fmt.Println("Error on ws write, exiting")
					ticker.Stop()
					return
				}
			}
		}
	}()

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
