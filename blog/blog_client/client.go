package main

import (
	"context"
	"io"
	"log"

	"github.com/Kaurin/gRPC/blog/blogpb"
	"google.golang.org/grpc"
)

func main() {
	log.Println("Blog Client started")

	opts := []grpc.DialOption{grpc.WithInsecure()}

	cc, err := grpc.Dial("localhost:50051", opts...)
	defer cc.Close()
	if err != nil {
		log.Fatalf("Could not connect: %v", err)
	}
	c := blogpb.NewBlogServiceClient(cc)

	//
	// CreateBlog
	//
	blog := &blogpb.Blog{
		// ID Creation handled by the server
		AuthorId: "Milos",
		Title:    "My first blog",
		Content:  "My content",
	}
	log.Println("Sending blog request...")
	createBlogResponse, err := c.CreateBlog(context.Background(), &blogpb.CreateBlogRequest{Blog: blog})
	if err != nil {
		log.Fatalf("Unexpected error: %v", err)
	}
	log.Printf("Blog has been created: %v", createBlogResponse)

	//
	// ReadBlog
	//
	log.Println("Reading the blog")

	// Invalid UUIDv4
	_, readBlogErr1 := c.ReadBlog(context.Background(), &blogpb.ReadBlogRequest{BlogId: "FORCEANERROR"})
	if readBlogErr1 != nil {
		log.Printf("Error happened while trying to read the blog: %v", readBlogErr1)
	}

	// Bogus UUID should throw an InvalidArgument error
	_, readBlogErr2 := c.ReadBlog(context.Background(), &blogpb.ReadBlogRequest{BlogId: "6b276f60-56cc-41bb-b0d5-cc9a94bd678c"})
	if readBlogErr2 != nil {
		log.Printf("Error happened while trying to read the blog: %v", readBlogErr2)
	}

	// Proper request
	readBlogReq, readBlogErr3 := c.ReadBlog(context.Background(), &blogpb.ReadBlogRequest{BlogId: createBlogResponse.GetBlog().GetId()})
	if readBlogErr3 != nil {
		log.Printf("Error happened while trying to read the blog: %v", readBlogErr3)
	}
	log.Printf("Got a response from the server: %v", readBlogReq)

	//
	// UpdateBlog
	//
	log.Println("Updating the blog")
	newBlog := &blogpb.Blog{
		Id:       createBlogResponse.GetBlog().GetId(),
		AuthorId: "MilosNumberTwo",
		Title:    "My first blog",
		Content:  "My content. Additional content.",
	}
	updateResp, updateErr := c.UpdateBlog(context.Background(), &blogpb.UpdateBlogRequest{Blog: newBlog})
	if updateErr != nil {
		log.Printf("Error happened while updating: %v", updateErr)
	}
	log.Printf("blog was updated: %v", updateResp)

	//
	// DeleteBlog
	//
	log.Println("Deleting the blog")

	// Incorrect UUID
	_, errDel := c.DeleteBlog(context.Background(), &blogpb.DeleteBlogRequest{BlogId: "BOGUS"})
	if errDel != nil {
		log.Printf("Yo, failed to delete blog: %v", errDel)
	}

	// non-existing blog
	_, errDel2 := c.DeleteBlog(context.Background(), &blogpb.DeleteBlogRequest{BlogId: "8494585d-5638-4ce7-b545-5974f4cdd5b0"})
	if errDel2 != nil {
		log.Printf("Yo, failed to delete blog: %v", errDel2)
	}

	// Properly delete
	respDel, errDel3 := c.DeleteBlog(context.Background(), &blogpb.DeleteBlogRequest{BlogId: createBlogResponse.GetBlog().GetId()})
	if errDel3 != nil {
		log.Printf("Yo, failed to delete blog: %v", errDel3)
	}
	log.Printf("Successfully deleted blog: %v", respDel)

	//
	// ListBlog
	//
	log.Println("Listing the blog")

	respList, errList := c.ListBlog(context.Background(), &blogpb.ListBlogRequest{})
	if errList != nil {
		log.Fatalf("Failed to recieve blogs: %v", errList)
	}
	for {
		res, err := respList.Recv()
		if err == io.EOF {
			break
		}
		if err != nil {
			log.Fatalf("Issue while getting messages via gRPC: %v", err)
		}
		log.Printf("Got blog: %v", res.GetBlog())
	}
}
