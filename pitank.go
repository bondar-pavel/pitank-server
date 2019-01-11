package main

import (
	"encoding/json"
	"fmt"
	"log"
	"time"

	"github.com/gorilla/websocket"
)

const (
	// Time allowed to write a message to the peer.
	writeWait = 10 * time.Second

	// Time allowed to read the next pong message from the peer.
	pongWait = 6 * time.Second

	// Send pings to peer with this period. Must be less than pongWait.
	pingPeriod = (pongWait * 9) / 10

	// Maximum message size allowed from peer.
	maxMessageSize = 512

	// Command Queue size.
	commandQueueSize = 5
)

// Pitank structure stores connected tank details
type Pitank struct {
	Name               string    `json:"name"`
	Status             string    `json:"status"`
	LastRegistration   time.Time `json:"last_registration"`
	LastDeregistration time.Time `json:"last_deregistration"`
	commandChan        chan interface{}
	conn               *websocket.Conn
}

func NewPitank(name string) *Pitank {
	return &Pitank{
		Name:   name,
		Status: "created",
	}
}

func (p *Pitank) SendCommand(cmd interface{}) {
	if p.commandChan == nil {
		fmt.Println("Error! Command channel is closed!")
		return
	}

	p.commandChan <- cmd
}

// Connect assigns websocket connection and initialize variables
func (p *Pitank) Connect(conn *websocket.Conn) {
	p.conn = conn
	p.commandChan = make(chan interface{}, commandQueueSize)

	p.Status = "connected"
	p.LastRegistration = time.Now()

	fmt.Printf(
		"Tank '%s' is connected at %s\n",
		p.Name,
		p.LastRegistration.Format(time.RFC3339))
}

// Disconnect closes websocket connection and command channel
func (p *Pitank) Disconnect() {
	if p.commandChan == nil {
		fmt.Println("Tank is already disconnected, cleaning up the rest")
		return
	}

	p.commandChan = nil
	p.conn.Close()

	p.Status = "disconnected"
	p.LastDeregistration = time.Now()

	fmt.Printf(
		"Tank '%s' is disconnected at %s\n",
		p.Name,
		p.LastDeregistration.Format(time.RFC3339))
}

// ReadPump currently used to receive websocket close event from client
// to start task deregistration
func (p *Pitank) ReadPump() {
	defer func() {
		fmt.Println("Disconnect from read")
		if p.commandChan != nil {
			close(p.commandChan)
		}
		fmt.Println("Disconnect from read done")
	}()
	p.conn.SetReadLimit(maxMessageSize)
	//p.conn.SetReadDeadline(time.Now().Add(pongWait))

	for {
		_, _, err := p.conn.ReadMessage()
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				log.Printf("error: %v", err)
			}
			break
		}
	}
}

// WritePump pumps commands for queue to websocket
func (p *Pitank) WritePump() {
	ticker := time.NewTicker(pingPeriod)
	defer func() {
		ticker.Stop()
		fmt.Println("Disconnect from write")
		p.Disconnect()
		fmt.Println("Disconnect from write done")
	}()
	for {
		select {
		case cmd, ok := <-p.commandChan:
			if !ok {
				// Command channel is closed, terminate connection
				fmt.Println("Command channel is closed, terminating connection")
				p.conn.WriteMessage(websocket.CloseMessage, []byte{})
				return
			}

			out, err := json.Marshal(cmd)
			if err != nil {
				fmt.Println("Error on marshal cmd:", err)
				return
			}
			fmt.Println("Sending command:", string(out))

			err = p.conn.WriteJSON(cmd)
			if err != nil {
				fmt.Println("Error on websocket write:", err)
				return
			}
		case <-ticker.C:
			p.conn.SetWriteDeadline(time.Now().Add(writeWait))
			if err := p.conn.WriteMessage(websocket.PingMessage, nil); err != nil {
				return
			}
		}
	}
}
