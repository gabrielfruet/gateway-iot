package main

import (
	"fmt"
	pb "gateway/proto"

	"github.com/google/uuid"
	amqp "github.com/rabbitmq/amqp091-go"
	"google.golang.org/protobuf/proto"
)

type Actuator struct {
    channel *amqp.Channel
    id uuid.UUID
    name string
    data string
    disconnect chan struct{}
}

func actuatorFromConnection(ch *amqp.Channel, sensor *pb.ConnectionRequest) (*Actuator, error) {
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

    return &Actuator{
        channel:     ch,
        id:          id,
        name:        sensor.GetQueueName(),
        data:        "",
        disconnect:  make(chan struct{}),
    }, nil
}
