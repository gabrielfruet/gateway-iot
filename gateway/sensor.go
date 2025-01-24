package main

import (
	"fmt"
	pb "gateway/proto"
	"log"
	"log/slog"

	uuid "github.com/google/uuid"
	amqp "github.com/rabbitmq/amqp091-go"
	"google.golang.org/protobuf/proto"
)

const DEVICE_TYPE_SENSOR = pb.DeviceType_DEVICE_TYPE_SENSOR
const DEVICE_TYPE_ACTUATOR = pb.DeviceType_DEVICE_TYPE_ACTUATOR

type Sensor struct {
    channel     *amqp.Channel
	id          uuid.UUID
    name        string
	data        string
    disconnect  chan struct{}
}

func (s *Sensor) GetQueueDelivery() (<-chan amqp.Delivery, error) {
	_, err := s.channel.QueueDeclare(
		s.name, // name
		false,            // durable
		false,            // delete when unused
		false,             // exclusive
		false,            // no-wait
		nil,              // arguments
	)


	if err != nil {
		return nil, err
	}

	delivery, err := s.channel.Consume(
		s.name, // queue
		"",               // consumer
		true,             // auto-ack
		false,             // exclusive
		false,            // no-local
		false,            // no-wait
		nil,              // args
	)

	if err != nil {
		return nil, err
	}

    return delivery, nil
}

func (s *Sensor) ListenUpdates() error {
    delivery, err := s.GetQueueDelivery()

    if err != nil {
        return err
    }

	for {
		select {
		case msg, ok := <-delivery:
			if !ok {
				log.Printf("Delivery channel closed for %s\n", s.name)
				return nil
			}

			sensor_data := pb.SensorDataUpdate{}
			if err := proto.Unmarshal(msg.Body, &sensor_data); err != nil {
				log.Printf("Failed to unmarshal message: %v\n", err)
				continue
			}

			if sensor_data.GetId() != s.id.URN() {
				slog.Warn(fmt.Sprintf("Received update with wrong ID at queue %s", s.name))
				continue
			}

			slog.Info(fmt.Sprintf("Received update from %s; DATA=%s\n", s.name, sensor_data.GetData()))

		case <-s.disconnect:
			slog.Info(fmt.Sprintf("Stopping %s from receiving updates\n", s.name))
			return nil
		}
	}
}

func sensorFromConnection(ch *amqp.Channel, sensor *pb.ConnectionRequest) (*Sensor, error) {
    id := uuid.New()

    idURN := id.URN()

    connectionResponse := pb.ConnectionResponse{
        Id: &idURN,
    }

    connectionResponseBytes, err := proto.Marshal(&connectionResponse)

	if err != nil {
		return nil, err
	}

    ch.Publish("", fmt.Sprintf("%s_id", sensor.GetQueueName()), false, false, amqp.Publishing{
        ContentType: "text/plain",
        Body:        connectionResponseBytes,
    })

	return &Sensor{
        channel:     ch,
		id:          id,
        name:        sensor.GetQueueName(),
		data:        "",
        disconnect:        make(chan struct{}),
	}, nil
}
