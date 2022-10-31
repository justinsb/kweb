FROM golang:1.19-bullseye

RUN apt-get update && apt-get install -y protobuf-compiler

RUN go install google.golang.org/protobuf/cmd/protoc-gen-go@v1.28.1 && mv /go/bin/protoc-gen-go /bin
RUN GO111MODULE=on GOBIN=/bin go install github.com/bufbuild/buf/cmd/buf@v1.9.0

WORKDIR /workspace
