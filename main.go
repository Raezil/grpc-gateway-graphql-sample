package main

import (
	"db"
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
	client := db.NewClient()
	if err := client.Prisma.Connect(); err != nil {
		panic(err)
	}

	defer func() {
		if err := client.Prisma.Disconnect(); err != nil {
			panic(err)
		}
	}()
	server := grpc.NewServer(
		grpc.UnaryInterceptor(backend.AuthUnaryInterceptor),
	)
	backend.RegisterGreeterServer(server, &backend.Server{
		PrismaClient: client,
	})
	server.Serve(conn)
}
