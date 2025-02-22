setup:
	go install google.golang.org/protobuf/cmd/protoc-gen-go@latest
	go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@latest
	docker network create gateway-iot

PROTO_SRC:=proto/
GO_DST:=gateway
PY_DST:=devices
PROTO_FILES:=proto/*.proto

PROTO_DST:=$(GO_DST) $(PY_DST)

.PHONY: protoc
protoc: go_protoc py_protoc

.PHONY: go_protoc
go_protoc: $(GO_DST)/proto

.PHONY: py_protoc
py_protoc: $(PY_DST)/proto

$(GO_DST)/proto: $(PROTO_FILES)
	protoc -I=. --go_opt=paths=source_relative --go_out=$(GO_DST) \
		--go-grpc_opt=paths=source_relative --go-grpc_out=$(GO_DST) $(PROTO_FILES)

$(PY_DST)/proto: $(PROTO_FILES)
	python -m grpc_tools.protoc -I=. --python_out=$(PY_DST) --pyi_out=$(PY_DST)  --grpc_python_out=$(PY_DST) $(PROTO_FILES)

.PHONY: rabbitmq
rabbitmq: 
	docker run --network gateway-iot -it --rm --name rabbitmq -p 5672:5672 -p 15672:15672 rabbitmq:4.0-management

.PHONY: gateway
gateway:
	cd gateway &&\
		docker build -t gateway-iot . &&\
		docker run -it --rm	--name gateway-iot \
			-p 8080:8080 \
			--network gateway-iot \
			gateway-iot \
			./gateway;

.PHONY: devices
devices:
	docker build -t devices-iot -f devices/Dockerfile . && \
		docker run -it --rm	--name devices-iot \
			--network gateway-iot \
			devices-iot \
			./launch.sh	



.PHONY: devices_bash
devices_bash:
	docker build -t devices-iot -f devices/Dockerfile . && \
		docker run -it --rm	--name devices-iot \
			--network gateway-iot \
			devices-iot \
			/bin/bash
