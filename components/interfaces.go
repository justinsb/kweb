package components

import (
	"context"
	"net/http"

	"github.com/justinsb/kweb/templates/scopes"
	"google.golang.org/protobuf/proto"
)

type Request struct {
	Session Session

	*http.Request

	PathParameters map[string]string
}

// Session implements session storage.
type Session interface {
	Clear(_ proto.Message)
	Set(value proto.Message)
	Get(dest proto.Message) bool
}

var contextKeyRequest = &Request{}

func GetRequest(ctx context.Context) *Request {
	return ctx.Value(contextKeyRequest).(*Request)
}

type Component interface {
	RegisterHandlers(server *Server, mux *http.ServeMux) error

	// AddToScope adds values to the template scope
	AddToScope(ctx context.Context, scope *scopes.Scope)
}

type RequestFilterChain func(ctx context.Context, req *Request) (Response, error)
type RequestFilterFunction func(ctx context.Context, req *Request, next RequestFilterChain) (Response, error)
type RequestFilter interface {
	ProcessRequest(ctx context.Context, req *Request, next RequestFilterChain) (Response, error)
}

func (r *Request) PathParameter(key string) string {
	return r.PathParameters[key]
}
