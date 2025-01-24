package main

import (
	"fmt"
	pb "gateway/proto"
	"log/slog"
	"sync"

	amqp "github.com/rabbitmq/amqp091-go"
	proto "google.golang.org/protobuf/proto"
)

type Gateway struct {
	devices    map[string]*Sensor
	ch         *amqp.Channel
	deviceLock sync.RWMutex
}

func newGateway() (*Gateway, error) {
	conn, err := amqp.Dial("amqp://guest:guest@localhost:5672/")

	if err != nil {
		return nil, err
	}

	ch, err := conn.Channel()

	if err != nil {
		return nil, err
	}

	return &Gateway{
		devices: make(map[string]*Sensor),
		ch:      ch,
	}, nil
}

func (g *Gateway) RemoveDevice(name string) {
    g.devices[name].disconnect <- struct{}{}

    g.deviceLock.Lock()
    slog.Debug(fmt.Sprintf("Deleting element from devices queue: %s", name))
    delete(g.devices, name)
    g.deviceLock.Unlock()
}


func (g *Gateway) ListenDisconnections() error {
	q, err := g.ch.QueueDeclare(
		"disconnect", // name
		false,     // durable
		false,     // delete when unused
		false,     // exclusive
		false,     // no-wait
		nil,       // arguments
	)

	if err != nil {
		return err
	}

    connection, err := g.ch.Consume(
		q.Name,    // queue
		"",        // consumer
		true,      // auto-ack
		false,     // exclusive
		false,     // no-local
		false,     // no-wait
		nil,       // args
	)


	slog.Info("Gateway is listening for device disconnections...")

	for msg := range connection {
		disconnectionRequest := &pb.DisconnectionRequest{}
		err := proto.Unmarshal(msg.Body, disconnectionRequest)

		if err != nil {
			slog.Error(fmt.Sprintf("Error when err unmarshalling disconection request: %v", err))
			continue
		}

        qname := disconnectionRequest.GetQueueName()

		slog.Info(fmt.Sprintf("Received disconnection request: queue: %v", qname))

        if _, ok := g.devices[qname]; !ok {
            slog.Info(fmt.Sprintf("Device with queue %s is not registered", qname))
            continue
        }
        g.RemoveDevice(qname)
	}

    return nil
}

func (g *Gateway) AddDevice(device *Sensor) {
    g.deviceLock.Lock()
    g.devices[device.name] = device
    g.deviceLock.Unlock()
}

func (g *Gateway) ListenConnections() error {
	q, err := g.ch.QueueDeclare(
		"connect", // name
		false,     // durable
		false,     // delete when unused
		false,     // exclusive
		false,     // no-wait
		nil,       // arguments
	)

	if err != nil {
		return err
	}

    connection, err := g.ch.Consume(
		q.Name,    // queue
		"",        // consumer
		true,      // auto-ack
		false,     // exclusive
		false,     // no-local
		false,     // no-wait
		nil,       // args
	)


	slog.Info("Gateway is listening for device connections...")

	for msg := range connection {
		connectionRequest := &pb.ConnectionRequest{}
		err := proto.Unmarshal(msg.Body, connectionRequest)

		if err != nil {
			slog.Error(fmt.Sprintf("Error when err unmarshalling connection request: %v", err))
			continue
		}

		slog.Info(fmt.Sprintf("Received connection request: queue: %v", connectionRequest.GetQueueName()))

        if v, ok := g.devices[connectionRequest.GetQueueName()]; ok {
            slog.Error(fmt.Sprintf("Device with queue %s already exists", connectionRequest.GetQueueName()))
            slog.Error(fmt.Sprintf("Ok was: %v and v was %v", ok, v))
            continue
        }


		device, err := sensorFromConnection(g.ch, connectionRequest)

        if err != nil {
            slog.Error(fmt.Sprintf("Error when creating device from connection: %v", err))
            continue
        }

        slog.Info(fmt.Sprintf("Device with queue %s was assigned to the uuid%s",
            connectionRequest.GetQueueName(), device.id.URN()))

        go g.AddDevice(device)

		go device.ListenUpdates()
	}

    return nil
}
