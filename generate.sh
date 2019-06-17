#!/bin/bash

set -x

protoc --go_out=plugins=grpc:. greet/greetpb/greet.proto
protoc --go_out=plugins=grpc:. calculator/calculatorpb/calculator.proto

