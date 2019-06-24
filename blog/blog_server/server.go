package main

import (
	"context"
	"fmt"
	"log"
	"net"
	"os"
	"os/signal"
	"strings"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"github.com/Kaurin/gRPC/blog/blogpb"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/aws/awserr"
	"github.com/aws/aws-sdk-go-v2/aws/external"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb/dynamodbattribute"
	uuid "github.com/satori/go.uuid"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
)

var ddbClient *dynamodb.Client

var blogTable = "blogTable" // Name of the DDB table

type server struct{}

func (*server) CreateBlog(ctx context.Context, req *blogpb.CreateBlogRequest) (*blogpb.CreateBlogResponse, error) {
	log.Printf("Started 'CreateBlog' func with the following input: %v", req)

	blogID := uuid.NewV4()
	blog := req.GetBlog()
	blog.Id = blogID.String()

	av, err := dynamodbattribute.MarshalMap(blog) // From DDB docos. You can marshal arbitrary structs as long as the ID format matches!
	if err != nil {
		panic(fmt.Sprintf("failed to DynamoDB marshal Record, %v", err))
	}

	ddbInput := &dynamodb.PutItemInput{
		TableName: &blogTable,
		Item:      av,
	}

	ddbReq := ddbClient.PutItemRequest(ddbInput)

	_, ddbErr := ddbReq.Send(context.Background()) // DDB Response is empty on success (or just gives API request ID). Discarding.
	if ddbErr != nil {
		return nil, status.Errorf( // PROPERLY RETURNING gRPC ERRORS!
			codes.Internal,
			fmt.Sprintf("Could not send to DynamoDB: %v", ddbErr),
		)
	}
	log.Printf("Successfully written to DDB!")

	log.Printf("Finished 'CreateBlog'. Returning (wrapped in a response struct): %v", blog)
	return &blogpb.CreateBlogResponse{
		Blog: blog,
	}, nil
}

func (*server) ReadBlog(ctx context.Context, req *blogpb.ReadBlogRequest) (*blogpb.ReadBlogResponse, error) {
	log.Printf("Started 'ReadBlog' func with the following input: %v", req)

	blogID := req.GetBlogId()

	// Check UUID for errors
	_, uuidErr := uuid.FromString(blogID)
	if uuidErr != nil {
		return nil, status.Errorf( // PROPERLY RETURNING gRPC ERRORS!
			codes.InvalidArgument,
			fmt.Sprintf("Blog ID Provided does not match UUIDv4 format: %v", uuidErr),
		)
	}

	// Craft DDB request input
	ddbInput := &dynamodb.GetItemInput{
		Key: map[string]dynamodb.AttributeValue{
			"id": {
				S: aws.String(blogID),
			},
		},
		TableName: aws.String(blogTable),
	}

	// Perform DDB Request
	ddbReq := ddbClient.GetItemRequest(ddbInput)
	ddbResp, ddbErr := ddbReq.Send(context.Background())
	if ddbErr != nil {
		return nil, status.Errorf( // PROPERLY RETURNING gRPC ERRORS!
			codes.Internal,
			fmt.Sprintf("Could not get Blog from DynamoDB: %v", ddbErr),
		)
	}

	blog := &blogpb.Blog{}
	dynamodbattribute.UnmarshalMap(ddbResp.Item, blog)

	// NOT FOUND error
	if blog.GetId() == "" {
		return nil, status.Errorf( // PROPERLY RETURNING gRPC ERRORS!
			codes.NotFound,
			fmt.Sprintf("Could not find Blog from DynamoDB for key: %v", blogID),
		)
	}
	log.Printf("Finished 'ReadBlog'. Returning (wrapped in a response struct): %v", blog)

	return &blogpb.ReadBlogResponse{
		Blog: blog,
	}, nil

}

// TODO: Fails when updating with empty strings. Read through the DDB docos to see how to handle that.
// Maybe just use `PutItem` with the attribute condition? Not sure if possible
func (*server) UpdateBlog(ctx context.Context, req *blogpb.UpdateBlogRequest) (*blogpb.UpdateBlogResponse, error) {
	log.Printf("Started 'UpdateBlog' func with the following input: %v", req)

	ddbCondition := "attribute_exists(id)" // Only update if ID existed in DDB table.

	blog := req.GetBlog()
	blogID := blog.GetId()
	blogAuthor := blog.GetAuthorId()
	blogContent := blog.GetContent()
	blogTitle := blog.GetTitle()

	// Check UUID for errors
	_, uuidErr := uuid.FromString(blogID)
	if uuidErr != nil {
		return nil, status.Errorf( // PROPERLY RETURNING gRPC ERRORS!
			codes.InvalidArgument,
			fmt.Sprintf("Blog ID Provided does not match UUIDv4 format: %v", uuidErr),
		)
	}

	// Craft DDB request input
	// Unfortunately, can't use dynamodb. marshal/unmarshal here :(
	input := &dynamodb.UpdateItemInput{
		ConditionExpression: aws.String(ddbCondition), // Only update if ID existed in DDB table.
		ExpressionAttributeNames: map[string]string{
			"#A": "author_id",
			"#C": "content",
			"#T": "title",
		},
		ExpressionAttributeValues: map[string]dynamodb.AttributeValue{
			":a": {
				S: aws.String(blogAuthor),
			},
			":c": {
				S: aws.String(blogContent),
			},
			":t": {
				S: aws.String(blogTitle),
			},
		},
		Key: map[string]dynamodb.AttributeValue{
			"id": {
				S: aws.String(blog.GetId()),
			},
		},
		ReturnValues:     dynamodb.ReturnValueUpdatedOld,
		TableName:        aws.String(blogTable),
		UpdateExpression: aws.String("SET #A = :a, #C = :c, #T = :t"),
	}

	// Perform DDB Request
	ddbReq := ddbClient.UpdateItemRequest(input)
	ddbResp, ddbErr := ddbReq.Send(context.Background())
	if ddbErr != nil {
		if aerr, ok := ddbErr.(awserr.Error); ok {
			switch aerr.Code() {
			case dynamodb.ErrCodeConditionalCheckFailedException:
				return nil, status.Errorf( // PROPERLY RETURNING gRPC ERRORS!
					codes.FailedPrecondition,
					fmt.Sprintf("Could not update Blog in DynamoDB. Failed DynamoDB PutItem Conditional: %v", ddbCondition),
				)
			default:
				return nil, status.Errorf( // PROPERLY RETURNING gRPC ERRORS!
					codes.Internal,
					fmt.Sprintf("Could not update Blog in DynamoDB: %v", ddbErr),
				)
			}
		} else {
			return nil, status.Errorf( // PROPERLY RETURNING gRPC ERRORS!
				codes.Internal,
				fmt.Sprintf("Could not update Blog in DynamoDB: %v", ddbErr),
			)
		}
	}

	log.Printf("Old values: %s", strings.ReplaceAll(ddbResp.String(), "\n", ""))

	log.Printf("Finished 'UpdateBlog'. Returning (wrapped in a response struct): %v", blog)
	return &blogpb.UpdateBlogResponse{
		Blog: blog,
	}, nil
}

func (*server) DeleteBlog(ctx context.Context, req *blogpb.DeleteBlogRequest) (*blogpb.DeleteBlogResponse, error) {
	log.Printf("Started 'DeleteBlog' func with the following input: %v", req)

	blogID := req.GetBlogId()

	// Check UUID for errors
	_, uuidErr := uuid.FromString(blogID)
	if uuidErr != nil {
		return nil, status.Errorf( // PROPERLY RETURNING gRPC ERRORS!
			codes.InvalidArgument,
			fmt.Sprintf("Blog ID Provided does not match UUIDv4 format: %v", uuidErr),
		)
	}

	// Craft DDB request input
	ddbInput := &dynamodb.DeleteItemInput{
		ReturnValues: "ALL_OLD",
		Key: map[string]dynamodb.AttributeValue{
			"id": {
				S: aws.String(blogID),
			},
		},
		TableName: aws.String(blogTable),
	}

	// Perform DDB Request
	ddbReq := ddbClient.DeleteItemRequest(ddbInput)
	ddbResp, ddbErr := ddbReq.Send(context.Background())
	if ddbErr != nil {
		return nil, status.Errorf( // PROPERLY RETURNING gRPC ERRORS!
			codes.Internal,
			fmt.Sprintf("Could not get delete from DynamoDB: %v", ddbErr),
		)
	}

	// NOT FOUND error
	if len(ddbResp.Attributes) == 0 {
		return nil, status.Errorf( // PROPERLY RETURNING gRPC ERRORS!
			codes.NotFound,
			fmt.Sprintf("Could not find Blog from DynamoDB for key: %v", blogID),
		)
	}
	log.Printf("Deleted blog from DDB: %s", strings.ReplaceAll(ddbResp.String(), "\n", ""))

	return &blogpb.DeleteBlogResponse{
		BlogId: blogID,
	}, nil
}

func (*server) ListBlog(req *blogpb.ListBlogRequest, stream blogpb.BlogService_ListBlogServer) error {
	input := &dynamodb.ScanInput{
		TableName: aws.String(blogTable),
	}

	// Example iterating over pages.
	ddbReq := ddbClient.ScanRequest(input)
	p := dynamodb.NewScanPaginator(ddbReq)

	for p.Next(context.Background()) {
		page := p.CurrentPage()
		log.Printf("%v", page)
		for _, item := range page.Items {
			blog := &blogpb.Blog{}
			dynamodbattribute.UnmarshalMap(item, blog)
			stream.Send(&blogpb.ListBlogResponse{
				Blog: blog,
			})
		}
	}

	if scanErr := p.Err(); scanErr != nil {
		log.Fatalf("Failed to paginate DynamoDB scan: %v", scanErr)
		return status.Errorf(codes.Internal,
			fmt.Sprintf("Failed to paginate DynamoDB scan: %v", scanErr),
		)
	}
	return nil
}

func localDynamoDB(ddbCfg aws.Config) aws.Config {
	ddbCfg.Credentials = aws.StaticCredentialsProvider{
		Value: aws.Credentials{
			AccessKeyID:     "BOGUS",
			SecretAccessKey: "BOGUS",
		},
	}
	ddbCfg.Region = "local"
	ddbCfg.EndpointResolver = aws.ResolveWithEndpointURL("http://dynamodb:8000")

	return ddbCfg
}

func main() {

	// If we crash the go code, we get the file name and line number
	log.SetFlags(log.LstdFlags | log.Lshortfile)
	log.Println("Blog program started...")

	lis, err := net.Listen("tcp", "0.0.0.0:50051")
	if err != nil {
		log.Fatalf("Failed to listen %v", err)
	}
	defer lis.Close()

	// AWS Dynamodb
	log.Println("Initializing DynamoDB Client")
	ddbCfg, ddbErr := external.LoadDefaultAWSConfig()
	if ddbErr != nil {
		panic("unable to load SDK config, " + ddbErr.Error())
	}

	if _, varSet := os.LookupEnv("LOCALDDB"); varSet { // If LOCALDDB is set (even if empty)
		ddbCfg = localDynamoDB(ddbCfg)
	}

	ddbClient = dynamodb.New(ddbCfg)

	ddbreq := ddbClient.CreateTableRequest(&dynamodb.CreateTableInput{
		TableName: aws.String(blogTable),
		AttributeDefinitions: []dynamodb.AttributeDefinition{
			dynamodb.AttributeDefinition{
				AttributeName: aws.String("id"),
				AttributeType: "S",
			},
		},
		KeySchema: []dynamodb.KeySchemaElement{
			dynamodb.KeySchemaElement{
				AttributeName: aws.String("id"),
				KeyType:       "HASH",
			},
		},
		ProvisionedThroughput: &dynamodb.ProvisionedThroughput{
			ReadCapacityUnits:  aws.Int64(1),
			WriteCapacityUnits: aws.Int64(1),
		},
	})

	ddbreq.Send(context.Background()) // I don't care if table creation works or not

	// gRPC server
	log.Printf("Registering gRPC server")
	opts := []grpc.ServerOption{}
	s := grpc.NewServer(opts...)
	defer s.Stop()

	// Register the reflection service on our gRPC server
	reflection.Register(s)

	// Register BlogServiceServer
	blogpb.RegisterBlogServiceServer(s, &server{})

	go func() {
		log.Println("Started Blog gRPC server in a separate GoRoutine")
		if err := s.Serve(lis); err != nil {
			log.Fatalf("Failed to serve: %v", err)
		}
	}()

	// Wait for Ctrl+c to exit
	ch := make(chan os.Signal, 1)
	signal.Notify(ch, os.Interrupt)

	// block until a signal is recieved
	<-ch
	fmt.Println("End of program!")
}
