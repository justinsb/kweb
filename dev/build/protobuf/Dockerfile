FROM golang:1.19-bullseye

RUN apt-get update && apt-get install -y protobuf-compiler

RUN go install google.golang.org/protobuf/cmd/protoc-gen-go@v1.28.1 && mv /go/bin/protoc-gen-go /bin

WORKDIR /workspace
