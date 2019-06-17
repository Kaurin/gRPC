package main

import (
	"context"
	"io"
	"log"
	"time"

	"google.golang.org/grpc/codes"

	"google.golang.org/grpc/status"

	"github.com/Kaurin/gRPC/calculator/calculatorpb"
	"google.golang.org/grpc"
)

func main() {
	// setup the client
	cc, err := grpc.Dial("localhost:50051", grpc.WithInsecure())
	defer cc.Close()
	if err != nil {
		log.Println("Can't establish gRPC connection.")
	}
	c := calculatorpb.NewCalculatorClient(cc)
	// doUnary(c)
	// doPrimeNumberDecomposition(c)
	// doComputeAverage(c)
	// doFindMaximum(c)
	doErrorUnary(c)
}

func doUnary(c calculatorpb.CalculatorClient) {
	log.Printf("Starting the Unary Client operation")
	// setup the request
	elems := &calculatorpb.AdditionElements{
		Elements: []int64{1, 3, 4},
	}
	req := &calculatorpb.SumRequest{
		SumElements: elems,
	}

	res, err := c.Sum(context.Background(), req)
	if err != nil {
		log.Fatalf("Unable to perform a gRPC call: %v", err)
	}
	log.Printf("Result: %v", res.GetResult())
}

func doPrimeNumberDecomposition(c calculatorpb.CalculatorClient) {
	log.Printf("Starting the Prime Number Decomposition operation")
	req := &calculatorpb.PNDRequest{
		Request: int64(9223372036854775806),
		// Request: int64(120),
	}
	recStream, err := c.PrimeNumberDecomposition(context.Background(), req)

	if err != nil {
		log.Fatalf("Unable to perform a gRPC call: %v", err)
	}

	for {
		res, err := recStream.Recv()
		if err == io.EOF {
			break
		}
		if err != nil {
			log.Fatalf("Issue while getting messages via gRPC: %v", err)
		}
		log.Printf("Got number: %v", res.GetResponse())
	}
}

func doComputeAverage(c calculatorpb.CalculatorClient) {
	log.Printf("Starting the Compute Average operation")
	stream, err := c.ComputeAverage(context.Background())
	if err != nil {
		log.Fatalf("Issue opening ComputeAverage gRPC: %v", err)
	}
	numbers := []int64{1, 3, 4, 67, 8, 12365, 35}
	for _, num := range numbers {
		log.Printf("Sending: %v", num)
		stream.Send(&calculatorpb.ComputeAverageRequest{
			Request: num,
		})
	}
	resp, err := stream.CloseAndRecv()
	if err != nil {
		log.Fatalf("Failed to Close/Recieve: %v", err)
	}
	log.Printf("Recieved average: %v", resp.GetAverage())
}

func doFindMaximum(c calculatorpb.CalculatorClient) {
	log.Printf("Starting the FindMaximum operation")
	waitc := make(chan struct{})
	stream, err := c.FindMaximum(context.Background())
	if err != nil {
		log.Fatalf("Error starting BiDi gRPC: %v", err)
	}
	go func() {
		numbers := []int64{5, -9223372036854775808, 1, 5, 3, 6, 2, 20}
		for _, number := range numbers {
			stream.Send(&calculatorpb.FindMaximumRequest{
				Number: number,
			})
			time.Sleep(100 * time.Millisecond)
		}
		err := stream.CloseSend()
		if err != nil {
			log.Fatalf("Failed to close stream after sending: %v", err)
		}
	}()
	go func() {
		for {
			resp, err := stream.Recv()
			if err == io.EOF {
				log.Printf("EOF from server. Closing.")
				break
			}
			if err != nil {
				log.Fatalf("Issue with recieving stream: %v", err)
				break
			}
			log.Printf("Recieved: %v", resp.GetCurrentMax())
		}
		close(waitc)
	}()
	<-waitc
}

func doErrorUnary(c calculatorpb.CalculatorClient) {
	log.Printf("Starting the doErrorUnary operation")

	// legitimate req
	doSquareRootCall(c, 10)
	doSquareRootCall(c, -5)

}

func doSquareRootCall(c calculatorpb.CalculatorClient, number int32) {
	log.Printf("Trying for square root of %v", number)
	res, err := c.SquareRoot(context.Background(), &calculatorpb.SquareRootRequest{
		Number: number,
	})
	if err != nil {
		respErr, ok := status.FromError(err)
		if ok {
			// actual error from gRPC (user error)
			log.Printf("gRPC Error message from server: %v", respErr.Message())
			log.Printf("gRPC Error code from server: %v", respErr.Code())
			if respErr.Code() == codes.InvalidArgument {
				log.Printf("We probably sent a negative number!")
			}
		} else {
			log.Fatalf("Big error calling SquareRoot: %v", err)
		}
	} else {
		log.Printf("Sqrt of %v is %v ", number, res.GetNumberRoot())
	}
}
