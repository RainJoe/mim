.PHONY: proto
proto:
    protoc -I pb/push pb/push/push.proto --go_out=plugins=grpc:pb/push
    protoc -I pb/logic pb/logic/logic.proto --go_out=plugins=grpc:pb/logic

.PHONY: gate
gate: proto
	go run cmd/gate/main.go --cfg="cmd/gate/config.toml"

.PHONY: logic
logic: proto
	go run cmd/logic/main.go --cfg="cmd/logic/config.toml"

.PHONY: build
build: proto
	CGO_ENABLED=false GOOS=linux GOARCH=amd64 go build -o gate cmd/gate/main.go
	CGO_ENABLED=false GOOS=linux GOARCH=amd64 go build -o logic cmd/logic/main.go

.PHONY: clean
clean:
	rm -f gate
	rm -f logic
