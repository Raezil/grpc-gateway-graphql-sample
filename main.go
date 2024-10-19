package main

import (
	"log"
	"net"

	"backend"

	"google.golang.org/grpc"
)

func main() {
	conn, err := net.Listen("tcp", ":8080")
	if err != nil {
		log.Fatal(err)
	}
	defer conn.Close()

	server := grpc.NewServer(
		grpc.UnaryInterceptor(backend.AuthUnaryInterceptor),
	)
	backend.RegisterGreeterServer(server, &backend.Server{})
	server.Serve(conn)
}
