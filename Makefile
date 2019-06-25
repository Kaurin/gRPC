# Can't use it because protobuf needs to be set up first, but then I run into a circular dependancy :(
# PKGS := $(shell go list ./...)

PKGS := \
	github.com/Kaurin/gRPC/blog/blog_client \
	github.com/Kaurin/gRPC/blog/blog_server \
	github.com/Kaurin/gRPC/calculator/calculator_client \
	github.com/Kaurin/gRPC/calculator/calculator_server \
	github.com/Kaurin/gRPC/greet/greet_client \
	github.com/Kaurin/gRPC/greet/greet_server


evans:
	cd /tmp ; git clone https://github.com/ktr0731/evans.git
	cd /tmp/evans ;	go install # Requires golang 1.12+, and a properly set-up GOHOME and GOROOT
	rm -rf /tmp/evans

clean:
	rm -rf ssl/server.*
	rm -rf ssl/ca.*
	find . -name '*.pb.go' -type f -exec rm {} \;
	rm -rf vendor

goclean:
	go clean -cache -testcache -i -x -modcache $(PKGS)

cleanimages:
	docker-compose down
	docker rmi golangrpc || true
	docker rmi docker.io/amazon/dynamodb-local || true
	docker rmi golang:alpine || true
	docker image prune -f || true

prep:
	go get -u github.com/golang/protobuf/protoc-gen-go
	go get -u google.golang.org/grpc
	make protobuf
	cd ssl ; sh genssl.sh

protobuf:
	protoc --go_out=plugins=grpc:. greet/greetpb/greet.proto
	protoc --go_out=plugins=grpc:. calculator/calculatorpb/calculator.proto
	protoc --go_out=plugins=grpc:. blog/blogpb/blog.proto

test:
	docker cp grpc_greet_1:/code/ssl/server.crt ssl/server.crt
	go run github.com/Kaurin/gRPC/greet/greet_client
	go run github.com/Kaurin/gRPC/calculator/calculator_client
	go run github.com/Kaurin/gRPC/blog/blog_client
	echo End of test!

lint:
	go fmt $(PKGS)
	go vet $(PKGS)

all: goclean clean prep lint

.PHONY: prep clean lint protobuf all test cleanimages goclean evans

