package main

import (
	"context"
	"io"
	"log"
	"time"

	"github.com/Kaurin/gRPC/greet/greetpb"

	"google.golang.org/grpc"
)

func main() {
	log.Println("hello")
	cc, err := grpc.Dial("localhost:50051", grpc.WithInsecure())
	defer cc.Close()
	if err != nil {
		log.Fatalf("Could not connect: %v", err)
	}
	c := greetpb.NewGreetServiceClient(cc)
	// doUnary(c)
	// doServerStreaming(c)
	// doClientStreaming(c)
	doBiDiStreaming(c)
}

func doUnary(c greetpb.GreetServiceClient) {
	log.Println("Starting the 'doUnary' RPC...")
	req := &greetpb.GreetRequest{
		Greeting: &greetpb.Greeting{
			FirstName: "John",
			LastName:  "Doe",
		},
	}

	res, err := c.Greet(context.Background(), req)
	if err != nil {
		log.Fatalf("Error while calling Greet RPC: %v", err)
	}

	log.Printf("Response from Greet: %v", res.GetResult())
}

func doServerStreaming(c greetpb.GreetServiceClient) {
	log.Println("Starting to do a server streaming RPC...")

	req := &greetpb.GreetManyTimesRequest{
		Greeting: &greetpb.Greeting{
			FirstName: "John",
			LastName:  "Doe",
		},
	}

	resStream, err := c.GreetManyTimes(context.Background(), req)
	if err != nil {
		log.Fatalf("Error while calling GreetManyTimes RPC: %v", err)
	}
	for {
		msg, err := resStream.Recv()
		if err == io.EOF {
			break
		}
		if err != nil {
			log.Fatalf("Error while reading stream: %v", err)
		}
		log.Printf("Response from GreetManyTimes: %v", msg.GetResult())
	}

}

func doClientStreaming(c greetpb.GreetServiceClient) {
	log.Println("Starting to do a client streaming RPC...")
	stream, err := c.LongGreet(context.Background())
	if err != nil {
		log.Fatalf("Error while calling LongGreet: %v", err)
	}

	requests := []*greetpb.LongGreetRequest{
		&greetpb.LongGreetRequest{
			Greeting: &greetpb.Greeting{
				FirstName: "John",
				LastName:  "Doe",
			},
		},
		&greetpb.LongGreetRequest{
			Greeting: &greetpb.Greeting{
				FirstName: "John1",
				LastName:  "Doe1",
			},
		},
		&greetpb.LongGreetRequest{
			Greeting: &greetpb.Greeting{
				FirstName: "John2",
				LastName:  "Doe2",
			},
		},
		&greetpb.LongGreetRequest{
			Greeting: &greetpb.Greeting{
				FirstName: "John3",
				LastName:  "Doe3",
			},
		},
		&greetpb.LongGreetRequest{
			Greeting: &greetpb.Greeting{
				FirstName: "John4",
				LastName:  "Doe4",
			},
		},
	}

	for _, req := range requests {
		log.Printf("Sending req: %v", req)
		stream.Send(req)
		time.Sleep(100 * time.Millisecond)
	}

	res, err := stream.CloseAndRecv()
	if err != nil {
		log.Fatalf("Error while recieving response from LongGreet: %v", err)
	}
	log.Printf("LongGreet Response: %v", res.GetResult())
}

func doBiDiStreaming(c greetpb.GreetServiceClient) {
	log.Println("Starting to do a BiDi streaming RPC...")

	stream, err := c.GreetEveryone(context.Background())
	if err != nil {
		log.Fatalf("Error while creating stream: %v", err)
	}

	requests := []*greetpb.GreetEveryoneRequest{
		&greetpb.GreetEveryoneRequest{
			Greeting: &greetpb.Greeting{
				FirstName: "John",
				LastName:  "Doe",
			},
		},
		&greetpb.GreetEveryoneRequest{
			Greeting: &greetpb.Greeting{
				FirstName: "John1",
				LastName:  "Doe1",
			},
		},
		&greetpb.GreetEveryoneRequest{
			Greeting: &greetpb.Greeting{
				FirstName: "John2",
				LastName:  "Doe2",
			},
		},
		&greetpb.GreetEveryoneRequest{
			Greeting: &greetpb.Greeting{
				FirstName: "John3",
				LastName:  "Doe3",
			},
		},
		&greetpb.GreetEveryoneRequest{
			Greeting: &greetpb.Greeting{
				FirstName: "John4",
				LastName:  "Doe4",
			},
		},
	}

	waitc := make(chan struct{})

	go func() {
		for _, req := range requests {
			log.Printf("Sending message: %v", req)
			stream.Send(req)
			time.Sleep(1 * time.Second)
		}
		stream.CloseSend()
	}()

	go func() {
		for {
			resp, err := stream.Recv()
			if err == io.EOF {
				break
			}
			if err != nil {
				log.Fatalf("Error while recieving: %v", err)
				break
			}
			log.Printf("Recieved: %v", resp.GetResult())
		}
		close(waitc)
	}()
	<-waitc
}
