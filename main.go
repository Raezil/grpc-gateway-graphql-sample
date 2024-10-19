package main

import (
	"context"
	"fmt"
	"log"
	"net"

	"backend"

	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"
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

func authUnaryInterceptor(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
	if info.FullMethod == "/authenticator.Auth/Login" || info.FullMethod == "/authenticator.Auth/Register" {
		return handler(ctx, req)
	}
	md, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		return nil, fmt.Errorf("missing metadata")
	}
	token := md["authorization"]
	if len(token) == 0 {
		return nil, fmt.Errorf("missing token")
	}

	claims, err := backend.VerifyJWT(token[0])
	ctx = metadata.AppendToOutgoingContext(ctx, "current_user", claims.Email)
	if err != nil {
		return nil, fmt.Errorf("unauthorized: %v", err)
	}
	return handler(ctx, req)
}

func main() {
	conn, err := net.Listen("tcp", ":8080")
	if err != nil {
		log.Fatal(err)
	}
	defer conn.Close()

	server := grpc.NewServer(
		grpc.UnaryInterceptor(authUnaryInterceptor),
	)
	backend.RegisterGreeterServer(server, &Server{})
	server.Serve(conn)
}
