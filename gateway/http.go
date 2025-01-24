package main

import (
	"encoding/json"
	"log"
	"net/http"
)

type HttpServer struct {
    gateway *Gateway
    server *http.Server
}

func (h *HttpServer) GetActuators(w http.ResponseWriter, r *http.Request) http.ResponseWriter {
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

func (h *HttpServer) GetSensors(w http.ResponseWriter, r *http.Request) http.ResponseWriter {
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

func (h *HttpServer) actuatorsHandler(w http.ResponseWriter, r *http.Request) {
    if r.Method == http.MethodGet {
        h.GetActuators(w,r)
    }
}

func (h *HttpServer) sensorsHandler(w http.ResponseWriter, r *http.Request) {
    if r.Method == http.MethodGet {
        h.GetSensors(w,r)
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
