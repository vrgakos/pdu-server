package main

import (
	"net"
	"fmt"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
	pb "pdu-server/protos"
	"pdu-server/app"
)



func grpcMain(mainApp *app.App) {
	listen, err := net.Listen("tcp", ":4000")
	if err != nil {
		fmt.Printf("failed to listen: %v\n", err)
		return
	}

	grpcServer := grpc.NewServer()
	pb.RegisterPduServerServer(grpcServer, mainApp)
	reflection.Register(grpcServer)
	grpcServer.Serve(listen)
}
