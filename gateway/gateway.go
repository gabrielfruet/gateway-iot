package main

import (
	"fmt"
	pb "gateway/proto"
	"log/slog"
	"sync"

	"github.com/google/uuid"
	amqp "github.com/rabbitmq/amqp091-go"
	proto "google.golang.org/protobuf/proto"
)

type DeviceType int

type Gateway struct {
    sensors    map[string]*Sensor
    actuators  map[string]*Actuator
    idType    map[uuid.UUID]DeviceType
    ch         *amqp.Channel
    sensorLock sync.RWMutex
    actuatorLock sync.RWMutex
    idLock sync.RWMutex
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
        sensors: make(map[string]*Sensor),
        actuators: make(map[string]*Actuator),
        idType : make(map[uuid.UUID]DeviceType),
        ch:      ch,
	}, nil
}

func (g *Gateway) GetActuators() []string {
    return getMapKeys(g.actuators)
}

func (g *Gateway) GetSensors() []string {
    return getMapKeys(g.sensors)
}

func (g *Gateway) RemoveSensor(disconnectRequest *pb.DisconnectionRequest) error {
    name := disconnectRequest.GetQueueName()
    id, err := uuid.Parse(disconnectRequest.GetId())

    if err != nil {
        return err
    }

    g.sensors[name].disconnect <- struct{}{}

    g.sensorLock.Lock()
    defer g.sensorLock.Unlock()
    slog.Debug(fmt.Sprintf("Deleting element from devices queue: %s", name))
    delete(g.sensors, name)

    g.idLock.Lock()
    defer g.idLock.Unlock()
    delete(g.idType, id)

    return nil
}

func (g *Gateway) RemoveActuator(disconnectRequest *pb.DisconnectionRequest) error {
    name := disconnectRequest.GetQueueName()
    id, err := uuid.Parse(disconnectRequest.GetId())

    if err != nil {
        return err
    }

    g.actuators[name].disconnect <- struct{}{}

    g.actuatorLock.Lock()
    defer g.actuatorLock.Unlock()
    slog.Debug(fmt.Sprintf("Deleting element from devices queue: %s", name))
    delete(g.actuators, name)

    g.idLock.Lock()
    defer g.idLock.Unlock()
    delete(g.idType, id)

    return nil
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

        if _, ok := g.sensors[qname]; !ok {
            slog.Info(fmt.Sprintf("Device with queue %s is not registered", qname))
            continue
        }
       
        id, err := uuid.Parse(disconnectionRequest.GetId())
        if err != nil {
            slog.Error(fmt.Sprintf("Error when parsing id: %v", err))
            continue
        }

        if g.idType[id] == DeviceType(pb.DeviceType_DEVICE_TYPE_SENSOR) {
            go g.RemoveSensor(disconnectionRequest)
        } else if g.idType[id] == DeviceType(pb.DeviceType_DEVICE_TYPE_ACTUATOR) {
            go g.RemoveActuator(disconnectionRequest)
        }
	}

    return nil
}

func (g *Gateway) AddSensor(device *Sensor) {
    g.sensorLock.Lock()
    defer g.sensorLock.Unlock()
    g.sensors[device.name] = device
    g.idLock.Lock()
    defer g.idLock.Unlock()
    g.idType[device.id] = DeviceType(pb.DeviceType_DEVICE_TYPE_SENSOR)
}

func (g *Gateway) AddActuator(device *Actuator) {
    g.actuatorLock.Lock()
    defer g.actuatorLock.Unlock()
    g.actuators[device.name] = device
    g.idLock.Lock()
    defer g.idLock.Unlock()
    g.idType[device.id] = DeviceType(pb.DeviceType_DEVICE_TYPE_ACTUATOR)
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

        if v, ok := g.sensors[connectionRequest.GetQueueName()]; ok {
            slog.Error(fmt.Sprintf("Device with queue %s already exists", connectionRequest.GetQueueName()))
            slog.Error(fmt.Sprintf("Ok was: %v and v was %v", ok, v))
            continue
        }

        if connectionRequest.GetType() == pb.DeviceType_DEVICE_TYPE_SENSOR {
            go func(){
                sensor, err := sensorFromConnection(g.ch, connectionRequest)

                if err != nil {
                    slog.Error(fmt.Sprintf("Error when creating device from connection: %v", err))
                }

                go g.AddSensor(sensor)
                go sensor.ListenUpdates()
            }()
        } else if connectionRequest.GetType() == pb.DeviceType_DEVICE_TYPE_ACTUATOR {
            go func() {
                actuator, err := actuatorFromConnection(g.ch, connectionRequest)

                if err != nil {
                    slog.Error(fmt.Sprintf("Error when creating device from connection: %v", err))
                }
                go g.AddActuator(actuator)
            }()
        }
	}

    return nil
}
