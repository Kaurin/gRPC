# https://www.dockerheart.com/_/golang
FROM golang:alpine

WORKDIR /code
COPY . .
RUN apk update
RUN apk add \
    make \
    protobuf \
    git \
    openssl
RUN make clean prep
