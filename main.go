package main

import (
	"context"
	"db"
	"fmt"
	"log"
	"net"

	"backend"

	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"
)

type Server struct {
	backend.UnimplementedGreeterServer
	PrismaClient *db.PrismaClient
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
func (s *Server) Login(ctx context.Context, in *backend.LoginRequest) (*backend.LoginReply, error) {
	log.Println("Login attempt for email:", in.Username)

	user, err := s.PrismaClient.User.FindUnique(
		db.User.Username.Equals(in.Username),
	).Exec(ctx)

	if err != nil {
		log.Printf("User not found: %v", err)
		return nil, fmt.Errorf("incorrect email or password")
	}

	if user.Password != in.Password {
		log.Println("Invalid password")
		return nil, fmt.Errorf("incorrect email or password")
	}

	token, err := backend.GenerateJWT(in.Username)
	if err != nil {
		log.Printf("Error generating token: %v", err)
		return nil, fmt.Errorf("could not generate token: %v", err)
	}

	log.Printf("Generated token: %s", token)

	return &backend.LoginReply{
		Token:   token,
		Message: "User was signed up!",
	}, nil
}

func (s *Server) SignUp(ctx context.Context, in *backend.SignUpRequest) (*backend.SignUpReply, error) {
	obj, err := s.PrismaClient.User.CreateOne(
		db.User.Password.Set(in.Password),
		db.User.Username.Set(in.Username),
	).Exec(ctx)

	if err != nil {
		log.Printf("failed to create user: %v", err)
		return nil, fmt.Errorf("failed to register user")
	}

	return &backend.SignUpReply{
		UserId:  obj.ID,
		Message: "User was created!",
	}, nil
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
