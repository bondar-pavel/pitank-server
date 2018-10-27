package main

import (
	"flag"
)

func main() {
	port := flag.String("port", "8080", "server port")
	flag.Parse()

	server := NewPitankServer(*port)
	server.Serve()
}
