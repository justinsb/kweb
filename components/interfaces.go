package components

import (
	"context"
	"net/http"

	"google.golang.org/protobuf/proto"
)

type Request struct {
	Session

	*http.Request
}

// Session implements session storage.
type Session interface {
	Clear(key string)
	Set(key string, value proto.Message)
	SetString(key string, value string)
	Get(key string) proto.Message
	GetString(key string) string
}

var contextKeyRequest = &Request{}

func GetRequest(ctx context.Context) *Request {
	return ctx.Value(contextKeyRequest).(*Request)
}

type Component interface {
	RegisterHandlers(server *Server, mux *http.ServeMux) error
	// ScopeValues are the values that are exposed to templates etc
	ScopeValues() any
	// Key is the unique name for this component, as used in templates etc
	Key() string
}

type RequestFilterChain func(ctx context.Context, req *Request) (Response, error)
type RequestFilterFunction func(ctx context.Context, req *Request, next RequestFilterChain) (Response, error)
type RequestFilter interface {
	ProcessRequest(ctx context.Context, req *Request, next RequestFilterChain) (Response, error)
}
