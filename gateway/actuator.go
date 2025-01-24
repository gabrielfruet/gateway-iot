package main

import "github.com/google/uuid"

type Actuator struct {
    id uuid.UUID
    name string
    data string
    disconnect chan struct{}
}
