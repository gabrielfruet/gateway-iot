package main

import (
	"log"
)

func main() {
	gateway, err := newGateway()

	if err != nil {
		log.Panic(err)
	}

    go gateway.ListenDisconnections()
	gateway.ListenConnections()
}
