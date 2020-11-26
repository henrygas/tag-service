package main

import (
	"google.golang.org/grpc"
	"log"
	"net"
	pb "tag-service/proto"
	"tag-service/server"
)

func main() {
	s := grpc.NewServer()
	pb.RegisterTagServiceServer(s, server.NewTagServer())

	listener, err := net.Listen("tcp", ":"+port)
	if err != nil {
		log.Fatalf("net.Listen err: %v", err)
	}

	err = s.Serve(listener)
	if err != nil {
		log.Fatalf("server.Serve err: %v", err)
	}
}