package main

import (
	"flag"
)

func main() {
	port := flag.String("port", "80", "server port")

	server := NewPitankServer(*port)
	server.Serve()
}
