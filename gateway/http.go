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
    response_body, err := json.Marshal(actuatorsNames)

    if err != nil {
        http.Error(w, err.Error(), http.StatusInternalServerError)
        return w
    }

    w.Write(response_body)
    return w
}

func (h *HttpServer) GetActuatorData(name string, w http.ResponseWriter) http.ResponseWriter {
    actuatorData, err := h.gateway.GetActuatorData(name)

    if err != nil {
        http.Error(w, err.Error(), http.StatusNotFound)
        return w
    }

    log.Printf("Actuator data: %v", actuatorData)
    response_body, err := json.Marshal(actuatorData)

    if err != nil {
        http.Error(w, err.Error(), http.StatusInternalServerError)
        return w
    }

    w.Write(response_body)
    return w
}

func (h *HttpServer) GetSensors(w http.ResponseWriter) http.ResponseWriter {
    sensorsNames := h.gateway.GetSensors()
    log.Printf("Sensors: %v", sensorsNames)
    response_body, err := json.Marshal(sensorsNames)

    if err != nil {
        http.Error(w, err.Error(), http.StatusInternalServerError)
        return w
    }

    w.Write(response_body)
    return w
}

func (h *HttpServer) GetSensorData(name string, w http.ResponseWriter) http.ResponseWriter {
    sensorData, err := h.gateway.GetSensorData(name)

    if err != nil {
        http.Error(w, err.Error(), http.StatusNotFound)
        return w
    }

    log.Printf("Sensors data: %v", sensorData)
    response_body, err := json.Marshal(sensorData)

    if err != nil {
        http.Error(w, err.Error(), http.StatusInternalServerError)
        return w
    }

    w.Write(response_body)
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
    data, err := h.gateway.ChangeActuatorState(changeStatePayload.Name, changeStatePayload.State)

    if err != nil {
        http.Error(w, err.Error(), http.StatusBadRequest)
        return w
    }

    changeStatePayload.State = data

    response_body, err := json.Marshal(changeStatePayload)

    if err != nil {
        http.Error(w, err.Error(), http.StatusInternalServerError)
        return w
    }

    w.Write(response_body)
    return w
}

type ChangeStatePayload struct {
    Name string `json:"name"`
    State string `json:"state"`
}

func (h *HttpServer) actuatorsHandler(w http.ResponseWriter, r *http.Request) {
    w.Header().Set("Content-Type", "application/json")
    name := r.URL.Query().Get("name")
    if r.Method == http.MethodGet && name == "" {
        h.GetActuators(w)
    } else if r.Method == http.MethodGet {
        h.GetActuatorData(name, w)
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

func corsMiddleware(next http.Handler) http.Handler {
    return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        w.Header().Set("Access-Control-Allow-Origin", "*")
        w.Header().Set("Access-Control-Allow-Methods", "GET, POST, OPTIONS")
        w.Header().Set("Access-Control-Allow-Headers", "Content-Type")

        if r.Method == http.MethodOptions {
            w.WriteHeader(http.StatusNoContent)
            return
        }

        next.ServeHTTP(w, r)
    })
}

func (h *HttpServer) Start() {
    mux := http.NewServeMux()
    mux.HandleFunc("/sensors", h.sensorsHandler)
    mux.HandleFunc("/actuators", h.actuatorsHandler)

    handler := corsMiddleware(mux)
    h.server.Handler = handler
    h.server.ListenAndServe()
}

func NewHttpServer(gateway *Gateway) *HttpServer {
    server := &http.Server{
        Addr: ":8080",
    }

    return &HttpServer{
        server: server,
        gateway: gateway,
    }
}
