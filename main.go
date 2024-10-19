package main

import (
	"context"
	"fmt"
	"log"
	"net"

	"backend"

	"google.golang.org/grpc"
)

type Server struct {
	backend.UnimplementedGreeterServer
}

func (s *Server) SayHello(ctx context.Context, req *backend.HelloRequest) (*backend.HelloReply, error) {
	return &backend.HelloReply{
		Message: fmt.Sprintf("Hello, %s!", req.GetName()),
	}, nil
}

func (s *Server) SayGoodbye(ctx context.Context, req *backend.GoodbyeRequest) (*backend.GoodbyeReply, error) {
	return &backend.GoodbyeReply{
		Message: fmt.Sprintf("Good-bye, %s!", req.GetName()),
	}, nil
}

func main() {
	conn, err := net.Listen("tcp", ":8080")
	if err != nil {
		log.Fatal(err)
	}
	defer conn.Close()

	server := grpc.NewServer()
	backend.RegisterGreeterServer(server, &Server{})
	server.Serve(conn)
}
