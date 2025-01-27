package main

import (
	"encoding/json"
	"fmt"
	"log"
	"log/slog"
	"net/http"
)

type HttpServer struct {
    gateway *Gateway
    server *http.Server
}

func (h *HttpServer) GetActuators(w http.ResponseWriter) http.ResponseWriter {
    actuatorsNames := h.gateway.GetActuators()
    log.Printf("Actuators: %v", actuatorsNames)
    b, err := json.Marshal(actuatorsNames)

    if err != nil {
        http.Error(w, err.Error(), http.StatusInternalServerError)
        return w
    }

    w.Write(b)
    return w
}

func (h *HttpServer) GetSensors(w http.ResponseWriter) http.ResponseWriter {
    sensorsNames := h.gateway.GetSensors()
    log.Printf("Sensors: %v", sensorsNames)
    b, err := json.Marshal(sensorsNames)

    if err != nil {
        http.Error(w, err.Error(), http.StatusInternalServerError)
        return w
    }

    w.Write(b)
    return w
}

func (h *HttpServer) GetSensorData(name string, w http.ResponseWriter) http.ResponseWriter {
    sensorData, err := h.gateway.GetSensorData(name)

    if err != nil {
        http.Error(w, err.Error(), http.StatusNotFound)
        return w
    }

    log.Printf("Sensors data: %v", sensorData)
    b, err := json.Marshal(sensorData)

    if err != nil {
        http.Error(w, err.Error(), http.StatusInternalServerError)
        return w
    }

    w.Write(b)
    return w
}

func (h *HttpServer) ChangeActuatorState(w http.ResponseWriter, r *http.Request) http.ResponseWriter {
    var changeStatePayload ChangeStatePayload
    slog.Info("Sending request to change Actuator State")
    err := json.NewDecoder(r.Body).Decode(&changeStatePayload)
    slog.Info(fmt.Sprintf("payload: %v",changeStatePayload))

    if err != nil {
        http.Error(w, err.Error(), http.StatusBadRequest)
        return w
    }
    h.gateway.ChangeActuatorState(changeStatePayload.Name, changeStatePayload.State)
    return w
}

type ChangeStatePayload struct {
    Name string `json:"name"`
    State string `json:"state"`
}

func (h *HttpServer) actuatorsHandler(w http.ResponseWriter, r *http.Request) {
    w.Header().Set("Content-Type", "application/json")
    if r.Method == http.MethodGet {
        h.GetActuators(w)
    } else if r.Method == http.MethodPost {
        h.ChangeActuatorState(w,r)
    }
}

func (h *HttpServer) sensorsHandler(w http.ResponseWriter, r *http.Request) {
    w.Header().Set("Content-Type", "application/json")
    name := r.URL.Query().Get("name")
    if name == "" && r.Method == http.MethodGet {
        h.GetSensors(w)
    } else if name != "" && r.Method == http.MethodGet {
        h.GetSensorData(name, w)
    }
}

func (h *HttpServer) Start() {
    http.HandleFunc("/sensors", h.sensorsHandler)
    http.HandleFunc("/actuators", h.actuatorsHandler)
    h.server.ListenAndServe()
}

func newHttpServer(gateway *Gateway) *HttpServer {
    server := &http.Server{
        Addr: ":8080",
    }

    return &HttpServer{
        server: server,
        gateway: gateway,
    }
}
