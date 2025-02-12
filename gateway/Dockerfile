FROM golang:1.23.5-bullseye

WORKDIR /app
COPY go.mod go.sum ./

RUN go mod download

COPY *.go ./
COPY proto ./proto/

RUN go build -o /gateway

EXPOSE 8080

CMD ["/gateway"]
