package backend

import (
	"context"
	"db"
	"fmt"
	"log"

	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"
)

type Server struct {
	UnimplementedGreeterServer
	PrismaClient *db.PrismaClient
}

func (s *Server) SayHello(ctx context.Context, req *HelloRequest) (*HelloReply, error) {
	return &HelloReply{
		Message: fmt.Sprintf("Hello, %s!", req.GetName()),
	}, nil
}

func (s *Server) SayGoodbye(ctx context.Context, req *GoodbyeRequest) (*GoodbyeReply, error) {
	return &GoodbyeReply{
		Message: fmt.Sprintf("Good-bye, %s!", req.GetName()),
	}, nil
}

func AuthUnaryInterceptor(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
	if info.FullMethod == "/greeter.Greeter/Login" || info.FullMethod == "/greeter.Greeter/SignUp" {
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

	claims, err := VerifyJWT(token[0])
	ctx = metadata.AppendToOutgoingContext(ctx, "current_user", claims.Email)
	if err != nil {
		return nil, fmt.Errorf("unauthorized: %v", err)
	}
	return handler(ctx, req)
}
func (s *Server) Login(ctx context.Context, in *LoginRequest) (*LoginReply, error) {
	log.Println("Login attempt for username:", in.Username)

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

	token, err := GenerateJWT(in.Username)
	if err != nil {
		log.Printf("Error generating token: %v", err)
		return nil, fmt.Errorf("could not generate token: %v", err)
	}

	log.Printf("Generated token: %s", token)

	return &LoginReply{
		Token:   token,
		Message: "User was logged in!",
	}, nil
}

func (s *Server) SignUp(ctx context.Context, in *SignUpRequest) (*SignUpReply, error) {
	obj, err := s.PrismaClient.User.CreateOne(
		db.User.Password.Set(in.Password),
		db.User.Username.Set(in.Username),
	).Exec(ctx)

	if err != nil {
		log.Printf("failed to create user: %v", err)
		return nil, fmt.Errorf("failed to register user")
	}

	return &SignUpReply{
		UserId:  obj.ID,
		Message: "User was created!",
	}, nil
}
