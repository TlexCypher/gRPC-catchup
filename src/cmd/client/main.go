package main

import (
	"bufio"
	"context"
	"errors"
	"fmt"
	"io"
	"log"
	"os"

	hellopb "mygrpc/pkg/grpc"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

var (
	scanner *bufio.Scanner
	client  hellopb.GreetingServiceClient
)

func main() {
	fmt.Println("start gRPC client")
	scanner = bufio.NewScanner(os.Stdin)

	address := "localhost:8080"
	conn, err := grpc.NewClient(
		address,

		grpc.WithTransportCredentials(insecure.NewCredentials()),
	)
	if err != nil {
		log.Fatal("Connection failed.")
		return
	}
	defer conn.Close()

	client = hellopb.NewGreetingServiceClient(conn)

	for {
		fmt.Println("1: send Request")
		fmt.Println("2: Hello ServerStream")
		fmt.Println("3: Hello ClientStream")
		fmt.Println("4: Hello BiStream")
		fmt.Println("5: exit")
		fmt.Print("please enter >")

		scanner.Scan()
		in := scanner.Text()

		switch in {
		case "1":
			Hello()

		case "2":
			HelloServerStream()
		case "3":
			HelloClientStream()
		case "4":
			HelloBiStream()
		case "5":
			fmt.Println("bye.")
			goto M
		}
	}
M:
}

func HelloBiStream() {
	clientStream, err := client.HelloBiStreams(context.Background())
	if err != nil {
		fmt.Println(err)
		return
	}
	recv, send, count, currentCount := false, false, 5, 0
	for !(recv && send) {
		if !send {
			scanner = bufio.NewScanner(os.Stdin)
			scanner.Scan()
			name := scanner.Text()
			if err := clientStream.Send(&hellopb.HelloRequest{
				Name: name,
			}); err != nil {
				fmt.Println(err)
				send = true
			}
			currentCount++
			if currentCount == count {
				send = true
				if err := clientStream.CloseSend(); err != nil {
					fmt.Println(err)
				}
			}
		}
		if !recv {
			resp, err := clientStream.Recv()
			if err != nil {
				if !errors.Is(err, io.EOF) {
					fmt.Println(err)
				}
				recv = true
			} else {
				fmt.Println(resp.GetMessage())
			}
		}
	}
}

func HelloClientStream() {
	clientStream, err := client.HelloClientStream(context.Background())
	if err != nil {
		fmt.Println(err)
	}

	pushCount := 5
	fmt.Printf("Please enter %d names.\n", pushCount)
	for _ = range pushCount {
		scanner.Scan()
		name := scanner.Text()
		if err := clientStream.Send(&hellopb.HelloRequest{
			Name: name,
		}); err != nil {
			fmt.Println(err)
		}
	}
	resp, err := clientStream.CloseAndRecv()
	if err != nil {
		fmt.Println(err)
	} else {
		fmt.Println(resp.GetMessage())
	}
}

func HelloServerStream() {
	fmt.Println("Please enter your name.")
	scanner.Scan()
	name := scanner.Text()

	serverStream, err := client.HelloServerStream(context.Background(),
		&hellopb.HelloRequest{
			Name: name,
		})
	if err != nil {
		fmt.Println(err)
	}

	for {
		res, err := serverStream.Recv()
		if errors.Is(err, io.EOF) {
			fmt.Println("This is the end of server stream")
			break
		}
		if err != nil {
			fmt.Println(err)
		}
		fmt.Println(res)
	}
}

func Hello() {
	fmt.Println("Please enter your name.")
	scanner.Scan()
	name := scanner.Text()

	req := &hellopb.HelloRequest{
		Name: name,
	}
	res, err := client.Hello(context.Background(), req)
	if err != nil {
		fmt.Println(err)
	} else {
		fmt.Println(res.GetMessage())
	}
}
