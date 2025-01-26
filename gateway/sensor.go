package main

import (
	pb "gateway/proto"
	"sync"

	uuid "github.com/google/uuid"
)

type Sensor struct {
	id          uuid.UUID
    name        string
	data        string
    dataLock    sync.RWMutex
}

func (s *Sensor) SetData(data string) {
    s.dataLock.Lock()
    defer s.dataLock.Unlock()
    s.data = data
}

func (s *Sensor) GetData() string {
    s.dataLock.RLock()
    defer s.dataLock.RUnlock()
    return s.data
}

func newSensor(sensor* pb.SensorDataUpdate) (*Sensor, error) {
	return &Sensor{
		id:          uuid.New(),
        name:        sensor.GetName(),
		data:        sensor.GetData(),
	}, nil
}
