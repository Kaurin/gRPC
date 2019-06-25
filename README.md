# gRPC practice

Done while following this guide Udemy course: https://www.udemy.com/grpc-golang/

### What's in this project

Three different go packages. All have to do with learning gRPC and are independent from each other

##### greet
* Demoes types of gRPC communication (check it's proto file)
* Uses SSL unlike the other two services
* Proper gRPC error handling examples
* Proper deadline examples

##### calculator
* Lessons learned from greet. Where greet was instructor lead, calculator was meant for students to figure out their own solution
* Not sure if it has proper eror/deadline examples. I might have implemented some.

##### blog
* CRUD demonstration with a "blog" app.
* Instructor lead, but I deviated and used DynamoDB.
* If you are using my `docker-compose.yml`, there will be no need to worry about AWS credentials and region setup as I'm using a "dynamodb-local" image to provide the DynamoDB functionality.
* Only of the three that uses DynamoDB
* Not sure if it has proper eror/deadline examples. I might have implemented some.

### Setup:
Requires:
* docker
* docker-compose
* Optional (if you want to mess around with files directly without docker-compose)
    * GoLang 1.12+
    * Properly set-up GoLang environment (GOROOT/GOHOME)
    * make

### Usage

##### Spin up the servers

```bash
docker-compose up
```

##### In a separate tab, run clients
Skip this if you just want to explore with Evans
```bash
## GREET

# Needed for greet_client because that has an SSL example, and the cert is self-signed
docker cp grpc_greet_1:/code/ssl/server.crt ssl/server.crt
go run github.com/Kaurin/gRPC/greet/greet_client

## CALCULATOR
go run github.com/Kaurin/gRPC/calculator/calculator_client

### BLOG
go run github.com/Kaurin/gRPC/blog/blog_client
```

##### Exploring the gRPC API manually with Evans

```bash
## Unfortunately, Evans doesn't support a simple "go get -u" :(
MYDIR=$(pwd)
cd /tmp
git clone https://github.com/ktr0731/evans.git
cd evans
go install # Requires golang 1.12+, and a properly set-up GOHOME and GOROOT
cd "$MYDIR" # Or just go back to the project folder

## GREET
# Same as above, needed because it's a self-signed cert
docker cp grpc_greet_1:/code/ssl/server.crt ssl/server.crt
evans -p 50052 --tls --host localhost --cacert ssl/server.crt --reflection

## CALCULATOR
evans -p 50053 -r

## BLOG
evans -p 50051 -r
```
What you want in all three is to use `call <functionName>`. 

Evans has tab-compoletion, but if you want to see all functions in a given run, then use `show rpc`.

Also, you can use `ctrl+d` to end sending messages in examples that require it, or to just bail on input.


### Cleanup
Don't forget to run:
```bash
docker-compose down
```

##### Manual cleanup of remaining files

**WARNING** Please make sure to read the makefile before running the command

```bash
make cleanimages
```

