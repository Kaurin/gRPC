package main

import (
	"context"
	"fmt"
	"io"
	"log"
	"math"
	"net"

	"google.golang.org/grpc/codes"

	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
	"google.golang.org/grpc/status"

	"github.com/Kaurin/gRPC/calculator/calculatorpb"
)

type server struct{}

func main() {
	log.Println("Hello World!")
	lis, err := net.Listen("tcp", ":50051")
	if err != nil {
		log.Printf("Error setting up listener %v", err)
	}
	s := grpc.NewServer()
	calculatorpb.RegisterCalculatorServiceServer(s, &server{})

	// Register the reflection service on our gRPC server
	reflection.Register(s)

	if err := s.Serve(lis); err != nil {
		log.Fatalf("Failed to serve: %v", err)
	}
}

func addArray(numbs ...int64) int64 {
	result := int64(0)
	for _, numb := range numbs {
		result += numb
	}
	return result
}

func (*server) Sum(ctx context.Context, in *calculatorpb.SumRequest) (*calculatorpb.SumResponse, error) {
	log.Printf("Started serving for elemenets: %v", in.GetSumElements().String())
	elements := in.GetSumElements().GetElements()

	result := &calculatorpb.SumResponse{
		Result: addArray(elements...),
	}

	return result, nil
}

func (*server) PrimeNumberDecomposition(in *calculatorpb.PNDRequest, stream calculatorpb.CalculatorService_PrimeNumberDecompositionServer) error {
	log.Printf("Started Prime Number Decomposition server streaming function")
	divisor := int64(2)
	n := in.GetRequest()
	for n > 1 {
		if n%divisor == 0 {
			res := &calculatorpb.PNDResponse{
				Response: divisor,
			}
			stream.SendMsg(res)
			n /= divisor
		} else {
			divisor++
		}
	}
	return nil
}

func (*server) ComputeAverage(stream calculatorpb.CalculatorService_ComputeAverageServer) error {
	log.Printf("Started ComputeAverage client streaming function")

	sum := int64(0)
	iterations := float64(0)
	for {
		req, err := stream.Recv()
		if err == io.EOF {
			response := float64(sum) / iterations
			log.Printf("Returning average: %v, and closing.", response)
			return stream.SendAndClose(&calculatorpb.ComputeAverageResponse{
				Average: response,
			})
		}
		if err != nil {
			log.Printf("Failed to recieve message from stream: %v", err)
		}
		sum += req.GetRequest()
		iterations++
	}
}

func (*server) FindMaximum(stream calculatorpb.CalculatorService_FindMaximumServer) error {
	log.Printf("Started FindMaximum BiDi streaming function")

	currentMax := *new(int64)
	index := 0
	for {
		index++
		req, err := stream.Recv()
		if err == io.EOF {
			break
		}
		if err != nil {
			return err
		}
		currentNumber := req.GetNumber()
		if currentNumber > currentMax || index == 1 {
			currentMax = currentNumber
			log.Printf("Detected new maximum: %v. Sending to client.", currentMax)
			stream.Send(&calculatorpb.FindMaximumResponse{
				CurrentMax: currentMax,
			})
		}
	}
	return nil
}

func (*server) SquareRoot(ctx context.Context, req *calculatorpb.SquareRootRequest) (*calculatorpb.SquareRootResponse, error) {
	num := req.GetNumber()
	if num < 0 {
		return nil, status.Errorf(
			codes.InvalidArgument,
			fmt.Sprintf("Recieved a negative number: %v", num))
	}
	return &calculatorpb.SquareRootResponse{
		NumberRoot: math.Sqrt(float64(num)),
	}, nil
}
