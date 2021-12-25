package main

import (
	"flag"

	"github.com/bondar-pavel/pitank-server/pkg/server"
)

func main() {
	port := flag.String("port", "8080", "server port")
	flag.Parse()

	s := server.NewPitankServer(*port)
	s.Serve()
}
