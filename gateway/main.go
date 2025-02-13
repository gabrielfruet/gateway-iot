package main

import (
	"log"
)

func main() {
	gateway, err := NewGateway()
    httpServer := NewHttpServer(gateway)

	if err != nil {
		log.Panic(err)
	}

    go httpServer.Start()
    go gateway.ListenActuatorRegistration()
    gateway.ListenSensorUpdates()
    defer gateway.Close()
}
