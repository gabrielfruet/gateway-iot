package main

import (
    "github.com/google/uuid"
    pb "gateway/proto"
)

type Actuator struct {
    id uuid.UUID
    name string
    data string
    disconnect chan struct{}
}

func actuatorFromConnection(sensor *pb.ConnectionRequest) *Actuator {
}
