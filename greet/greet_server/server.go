package main

import (
	"context"
	"io"
	"log"
	"net"
	"strconv"
	"time"

	"github.com/Kaurin/gRPC/greet/greetpb"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type server struct{}

func main() {
	log.Println("Hello world")

	lis, err := net.Listen("tcp", "0.0.0.0:50051")
	if err != nil {
		log.Fatalf("Failed to listen %v", err)
	}

	s := grpc.NewServer()
	greetpb.RegisterGreetServiceServer(s, &server{})

	if err := s.Serve(lis); err != nil {
		log.Fatalf("Failed to serve: %v", err)
	}
}

func (*server) Greet(ctx context.Context, req *greetpb.GreetRequest) (*greetpb.GreetResponse, error) {
	log.Printf("Now running the server 'Greet' function with: %v", req)
	firstname := req.GetGreeting().GetFirstName()
	lastname := req.GetGreeting().GetLastName()
	result := "Hello " + firstname + " " + lastname + "."
	res := greetpb.GreetResponse{
		Result: result,
	}
	return &res, nil
}

func (*server) GreetManyTimes(req *greetpb.GreetManyTimesRequest, stream greetpb.GreetService_GreetManyTimesServer) error {
	log.Println("Function 'GreetManyTimes' has been invoked")
	firstName := req.GetGreeting().GetFirstName()
	for i := 0; i < 10; i++ {
		result := "Hello " + firstName + " number " + strconv.Itoa(i)
		res := &greetpb.GreetManyTimesResponse{
			Result: result,
		}
		stream.Send(res)
		time.Sleep(1 * time.Second)
	}
	return nil
}

func (*server) LongGreet(stream greetpb.GreetService_LongGreetServer) error {
	log.Println("Function 'LongGreet' has been invoked")
	result := "Hello "
	for {
		req, err := stream.Recv()
		if err == io.EOF {
			log.Println("We have finished reading the client stream")
			resp := &greetpb.LongGreetResponse{
				Result: result,
			}

			return stream.SendAndClose(resp)
		}
		if err != nil {
			log.Fatalf("Error while recieving stream: %v", err)
		}

		name := req.GetGreeting().GetFirstName() + " " + req.GetGreeting().GetLastName()
		result += name + "! "
	}

}

func (*server) GreetEveryone(stream greetpb.GreetService_GreetEveryoneServer) error {
	log.Println("Function 'GreetEveryone' has been invoked")
	for {
		req, err := stream.Recv()
		if err == io.EOF {
			return nil
		}
		if err != nil {
			log.Fatalf("Error recieving client stream: %v", err)
		}
		firstName := req.GetGreeting().GetFirstName()
		result := "Hello " + firstName + "!"
		sendErr := stream.Send(&greetpb.GreetEveryoneResponse{
			Result: result,
		})
		if sendErr != nil {
			log.Fatalf("Failed to send data to client stream: %v", sendErr)
		}
	}
}

func (*server) GreetWithDeadline(ctx context.Context, req *greetpb.GreetWithDeadlineRequest) (*greetpb.GreetWithDeadlineResponse, error) {
	log.Printf("Now running the server 'GreetWithDeadline' function with: %v", req)
	for i := 0; i < 3; i++ {
		if ctx.Err() == context.Canceled {
			log.Printf("Client cancelled request!")
			return nil, status.Error(codes.Canceled, "The client cancelled the request")
		}
		time.Sleep(1 * time.Second)
	}
	firstname := req.GetGreeting().GetFirstName()
	lastname := req.GetGreeting().GetLastName()
	result := "Hello " + firstname + " " + lastname + "."
	res := &greetpb.GreetWithDeadlineResponse{
		Result: result,
	}
	return res, nil
}
