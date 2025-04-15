package main

import (
	"context"
	"errors"
	"fmt"
	"io"
	"log"
	"net"
	"os"
	"os/signal"
	"time"

	hellopb "mygrpc/pkg/grpc"

	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
)

type myServer struct {
	hellopb.UnimplementedGreetingServiceServer
}

func (s *myServer) Hello(ctx context.Context, req *hellopb.HelloRequest) (*hellopb.HelloResponse, error) {
	return &hellopb.HelloResponse{
		Message: fmt.Sprintf("Hello, %s!", req.GetName()),
	}, nil
}
func (s *myServer) HelloServerStream(req *hellopb.HelloRequest, srv grpc.ServerStreamingServer[hellopb.HelloResponse]) error {
	resCount := 5
	for i := range resCount {
		if err := srv.Send(&hellopb.HelloResponse{
			Message: fmt.Sprintf("[%d] Hello, %s!", i, req.GetName()),
		}); err != nil {
			return err
		}
		time.Sleep(time.Second * 1)
	}
	return nil
}

func (s *myServer) HelloClientStream(stream grpc.ClientStreamingServer[hellopb.HelloRequest, hellopb.HelloResponse]) error {
	nameList := make([]string, 0, 10)
	for {
		req, err := stream.Recv()
		if errors.Is(err, io.EOF) {
			message := fmt.Sprintf("Hello, %v!", nameList)
			return stream.SendAndClose(&hellopb.HelloResponse{
				Message: message,
			})
		}
		if err != nil {
			return err
		}
		nameList = append(nameList, req.GetName())
	}
}

/*
1つリクエストを受け取るたびに、1つリクエストを返す前提でかく
streamを貼っている最中に、全ての関数は一度だけしか呼ばれない
*/
func (s *myServer) HelloBiStreams(stream grpc.BidiStreamingServer[hellopb.HelloRequest, hellopb.HelloResponse]) error {
	for {
		req, err := stream.Recv()
		if errors.Is(err, io.EOF) {
			return nil
		}
		if err != nil {
			return err
		}
		fmt.Printf("BiStream: %v\n", req.GetName())
		message := fmt.Sprintf("Hello, %v!", req.GetName())
		if err := stream.Send(&hellopb.HelloResponse{
			Message: message,
		}); err != nil {
			return err
		}
	}
}

func NewMyServer() *myServer {
	return &myServer{}
}

func main() {
	port := 8080
	listener, err := net.Listen("tcp", fmt.Sprintf(":%v", port))
	if err != nil {
		panic(err)
	}
	s := grpc.NewServer()
	hellopb.RegisterGreetingServiceServer(s, NewMyServer())
	/* Register grpc server to use grpcurl*/
	reflection.Register(s)
	go func() {
		log.Printf("start gRPC server port: %v\n", port)
		s.Serve(listener)
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, os.Interrupt)
	<-quit
	log.Println("stopping gRPC server")
	s.GracefulStop()
}
