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
    conn         *amqp.Connection
    sensorLock sync.RWMutex
    actuatorLock sync.RWMutex
    idLock sync.RWMutex
    close chan struct{}
}

func getConnection() (*amqp.Connection, error) {
	conn, err := amqp.Dial("amqp://guest:guest@localhost:5672/")

	if err != nil {
		return nil, err
	}
    return conn, nil
}

func newGateway() (*Gateway, error) {
    conn, err := getConnection()

    if err != nil {
        return nil, err
    }

	return &Gateway{
        sensors: make(map[string]*Sensor),
        actuators: make(map[string]*Actuator),
        idType : make(map[uuid.UUID]DeviceType),
        conn:      conn,
	}, nil
}

func (g *Gateway) Close() {
    g.close <- struct {} {}
    g.conn.Close()
}

func (g *Gateway) GetActuators() []string {
    return getMapKeys(g.actuators)
}

func (g *Gateway) GetSensors() []string {
    return getMapKeys(g.sensors)
}

func (g *Gateway) GetSensorData(name string) (string, error) {
    g.sensorLock.RLock()
    defer g.sensorLock.RUnlock()
    if sensor, ok := g.sensors[name]; ok {
        return sensor.data, nil
    }
    return "", fmt.Errorf("Sensor with name %s not found", name)
}

func (g *Gateway) ChangeActuatorState(name string, state string) (string, error) {
    g.actuatorLock.RLock()
    defer g.actuatorLock.RUnlock()

    actuator, ok := g.actuators[name]

    if !ok {
        return "", fmt.Errorf("Actuator with name %s not found", name)
    }

    actuator.ChangeState(state)
    return actuator.data, nil
}

func (g *Gateway) RemoveSensor(disconnectRequest *pb.DisconnectionRequest) error {
    name := disconnectRequest.GetQueueName()
    id, err := uuid.Parse(disconnectRequest.GetId())

    if err != nil {
        return err
    }

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

    g.actuatorLock.Lock()
    defer g.actuatorLock.Unlock()
    slog.Debug(fmt.Sprintf("Deleting element from devices queue: %s", name))
    delete(g.actuators, name)

    g.idLock.Lock()
    defer g.idLock.Unlock()
    delete(g.idType, id)

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
    slog.Info(fmt.Sprintf("Actuator registered with name: %s", device.name))
}

func (g *Gateway) ListenActuatorRegistration() {
    exchangeName := "actuator_registration_order_exchange"
    ch, err := g.conn.Channel()

    if err != nil {
        slog.Error(fmt.Sprintf("Failed to open a channel: %v", err))
    }

    err = ch.ExchangeDeclare(
        exchangeName, // name
        "fanout",     // kind
        false,        // durable
        false,        // auto-delete
        false,        // internal
        false,        // no-wait
        nil,          // arguments
    )

    if err != nil {
        slog.Error(fmt.Sprintf("Failed to declare exchange: %v", err))
        return
    }

    slog.Info("Sending registration order")

    err = ch.Publish(
        exchangeName,
        "",
        false,
        false,
        amqp.Publishing{
            ContentType: "text/plain",
            Body: []byte(""),
        },
        )

    if err != nil {
        slog.Error(fmt.Sprintf("Failed to publish a message: %v", err))
    }

    queueName := "connect"

    q, err := ch.QueueDeclare(
        queueName, // name
        false,     // durable
        false,     // delete when unused
        false,     // exclusive
        false,     // no-wait
        nil,       // arguments
        )

    msgs, err := ch.Consume(
        q.Name, // queue
        "",     // consumer
        true,   // auto-ack
        false,  // exclusive
        false,  // no-local
        false,  // no-wait
        nil,    // arguments
        )

    if err != nil {
        slog.Error(fmt.Sprintf("Failed to consume a message: %v", err))
        return
    }

    // Process messages in a goroutine
    slog.Info("Gateway is listening for actuator registrations...")
    for { 
        select {
            case <-g.close:
                return
            case msg := <-msgs:
                go g.HandleActuatorRegistration(msg)
        }
    }
}

func (g *Gateway) HandleActuatorRegistration(msg amqp.Delivery) {
    var connectionRequest pb.ConnectionRequest
    proto.Unmarshal(msg.Body, &connectionRequest)

    name := connectionRequest.GetQueueName()

    actuator, ok := g.actuators[name]

    if ok {
        slog.Info(fmt.Sprintf("Actuator with name %s already exists, overriding it.", name))
    }

    actuator, err := newActuator(&connectionRequest)

    if err != nil {
        slog.Error(fmt.Sprintf("Failed to create actuator: %v", err))
    }

    g.AddActuator(actuator)
}

func (g *Gateway) ListenSensorUpdates() {
    ch, err := g.conn.Channel()

    if err != nil {
        slog.Error(fmt.Sprintf("Failed to open a channel: %v", err))
    }
    slog.Info("Gateway is listening for sensor updates...")
    queueName := "sensor_updates"

    q, err := ch.QueueDeclare(
        queueName, // name
        false,     // durable
        false,     // delete when unused
        false,     // exclusive
        false,     // no-wait
        nil,       // arguments
        )

    msgs, err := ch.Consume(
        q.Name, // queue
        "",     // consumer
        true,   // auto-ack
        false,  // exclusive
        false,  // no-local
        false,  // no-wait
        nil,    // arguments
        )

    if err != nil {
        slog.Error(fmt.Sprintf("Failed to consume a message: %v", err))
        return
    }

    // Process messages in a goroutine
    for { 
        select {
            case <-g.close:
                return
            case msg := <-msgs:
                slog.Info("Message arrived")
                go g.HandleUpdate(msg)
        }
    }
}

func (g *Gateway) HandleUpdate(msg amqp.Delivery) {
    var sensorUpdate pb.SensorDataUpdate
    proto.Unmarshal(msg.Body, &sensorUpdate)

    name := sensorUpdate.GetName()

    slog.Info(fmt.Sprintf("Received updates from %s", name))

    sensor, ok := g.sensors[name]

    if !ok {
        g.AddSensor(&Sensor{
            name:   name,
            data:   sensorUpdate.GetData(),
        })
        return
    }

    sensor.SetData(sensorUpdate.GetData())
}
