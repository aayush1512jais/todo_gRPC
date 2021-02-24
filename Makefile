proto-gen:
	rm -f pb/*.pb.go
	protoc -I proto proto/*.proto --go_out=pb --go-grpc_out=pb

run-server:
	go run cmd/server/main.go -port 8000

run-client:
	go run cmd/client/main.go -address 0.0.0.0:8000