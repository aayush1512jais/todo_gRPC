package main

import (
	"flag"
	"fmt"
	"log"
	"net"

	"gitlab.com/aayushjaiswal/meds/pb"
	"gitlab.com/aayushjaiswal/meds/service"
	"google.golang.org/grpc"
)

func main() {
	port := flag.Int("port", 8000, "the server port")
	flag.Parse()
	log.Println("server on port: ", *port)

	medicineStore := service.NewInMemoryMedicineStore()
	imageStore := service.NewDiskImageStore("img")
	medicineServer := service.NewMedicineServer(medicineStore, imageStore)

	grpcServer := grpc.NewServer()
	pb.RegisterMedicineServiceServer(grpcServer, medicineServer)

	address := fmt.Sprintf("0.0.0.0:%d", *port)
	listener, err := net.Listen("tcp", address)
	if err != nil {
		log.Fatal("cannot start server: ", err)
	}
	err = grpcServer.Serve(listener)
	if err != nil {
		log.Fatal("cannot start server: ", err)
	}
}
