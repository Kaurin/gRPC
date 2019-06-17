# gRPC practice

Done while following this guide Udemy course: https://www.udemy.com/grpc-golang/

### Setup:
1. Ensure you have GoLang 1.12+
1. Ensure `GOROOT` and `GOPATH` are set properly
1. Ensure `protoc` utility is available on the system
1. Protoc GoLang plugin: `go get -u github.com/golang/protobuf/protoc-gen-go`
1. Not sure if needed outside of modfile: `go get -u google.golang.org/grpc`
1. It's easier if this repo is **not** cloned in `GOPATH` because we are using `go mod`
1. Run `go mod tidy`
1. Check the `*.proto` files. Modify if so desired.
1. Run `./generate.sh`
1. Run the respective server, then client. For example: 
```bash
go run calculator/calculator_server/server.go
# <separate terminal>
go run calculator/calculator_client/client.go
```
