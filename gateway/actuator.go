package main

import (
	"context"
	"fmt"
	pb "gateway/proto"
	"log/slog"
	"sync"

	"github.com/google/uuid"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

type Actuator struct {
    id uuid.UUID
    name string
    data string
    dataLock sync.RWMutex
    client pb.ActuatorClient 
}

func (a* Actuator) ChangeState(state string) {
    id := a.id.URN()

    slog.Info("Sending gRPC call for changing the state of actuator %s to %s", a.name, state)

    r, err := a.client.ChangeState(
        context.Background(),
        &pb.ActuatorState{Id: &id, State: &state},
        )

    if err != nil {
        slog.Error(fmt.Sprintf("RPC ChangeState: %s", err.Error()))
    }

    a.dataLock.Lock()
    defer a.dataLock.Unlock()
    a.data = r.GetState()
}

func newActuator(actuator *pb.ConnectionRequest) (*Actuator, error) {
    if actuator.GetIp() == "" && actuator.GetPort() == "" {
        return nil, fmt.Errorf("%s: ip and port must be provided", actuator.GetQueueName())
    }

    conn, err := grpc.NewClient(
        fmt.Sprintf("%s:%s", actuator.GetIp(), actuator.GetPort()),
        grpc.WithTransportCredentials(insecure.NewCredentials()),
        )

    slog.Info(fmt.Sprintf("Connecting to %s:%s gRPC", actuator.GetIp(), actuator.GetPort()))

    if err != nil {
        return nil, err
    }

    client := pb.NewActuatorClient(conn)

    return &Actuator{
        client:      client,
        name:        actuator.GetQueueName(),
        id:          uuid.New(),
        data:        actuator.GetData(),
    }, nil
}
