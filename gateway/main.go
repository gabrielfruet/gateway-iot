package main

import (
	"log"
)

func main() {
	gateway, err := newGateway()
    httpServer := newHttpServer(gateway)

	if err != nil {
		log.Panic(err)
	}

    go gateway.ListenDisconnections()
    go httpServer.Start()
	gateway.ListenConnections()
}
