FROM golang:1.19-bullseye as builder

WORKDIR /src

COPY go.mod go.sum ./

RUN go mod download

# Prebuild some library dependencies to warm the cache
RUN CGO_ENABLED=0 go build -v \
  k8s.io/klog/v2 \
  k8s.io/client-go/rest/... \
  google.golang.org/protobuf/...

COPY *.go .
COPY components components
COPY server server
COPY debug debug
COPY templates templates
RUN CGO_ENABLED=0 go build -v \
  github.com/justinsb/kweb/...

COPY apps/sso/*.go apps/sso/
COPY apps/sso/pages apps/sso/pages

RUN CGO_ENABLED=0 go build -v -o /kweb-sso ./apps/sso

FROM gcr.io/distroless/static

COPY --from=builder /kweb-sso /kweb-sso

ENTRYPOINT ["/kweb-sso"]
