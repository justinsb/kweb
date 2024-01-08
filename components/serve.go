package components

import (
	"context"
	"fmt"
	"net/http"

	"github.com/justinsb/kweb/templates/scopes"
	"k8s.io/klog/v2"
)

type Server struct {
	Components []Component
}

var contextKeyServer = &Server{}

func GetServer(ctx context.Context) *Server {
	return ctx.Value(contextKeyServer).(*Server)
}

func WithServer(ctx context.Context, s *Server) context.Context {
	return context.WithValue(ctx, contextKeyServer, s)
}

func GetComponent[T any](ctx context.Context, dest *T) error {
	server := GetServer(ctx)
	return GetComponentFromServer(server, dest)
}

func MustGetComponent[T any](ctx context.Context) T {
	var t T
	if err := GetComponent[T](ctx, &t); err != nil {
		klog.Fatalf("error getting component %T: %v", t, err)
	}
	return t
}

func GetComponentFromServer[T any](s *Server, dest *T) error {
	var matches []T
	for _, component := range s.Components {
		match, ok := component.(T)
		if ok {
			matches = append(matches, match)
		}
	}
	if len(matches) == 0 {
		// for _, component := range s.Components {
		// 	klog.Infof("have component of type %T", component)
		// 	var t T
		// 	klog.Infof("want %T", t)
		// }
		return fmt.Errorf("component not found")
	}
	if len(matches) > 1 {
		return fmt.Errorf("multiple matching components found")
	}
	*dest = matches[0]
	return nil
}

type Response interface {
	WriteTo(ctx context.Context, w http.ResponseWriter)
}

func (s *Server) ServeHTTP(fn func(ctx context.Context, req *Request) (Response, error)) func(w http.ResponseWriter, r *http.Request) {
	// TODO: Can we cache / build once?
	var filters []RequestFilterFunction
	for _, component := range s.Components {
		if filter, ok := component.(RequestFilter); ok {
			filters = append(filters, filter.ProcessRequest)
		}
	}

	next := func(ctx context.Context, req *Request) (Response, error) {
		return fn(ctx, req)
	}
	for i := len(filters) - 1; i >= 0; i-- {
		filter := filters[i]
		captureNext := next
		invoke := func(ctx context.Context, req *Request) (Response, error) {
			return filter(ctx, req, captureNext)
		}
		next = invoke
	}

	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()

		req := &Request{
			Request: r,
		}
		req.PathParameters = make(map[string]string)

		ctx = context.WithValue(ctx, contextKeyRequest, req)
		ctx = context.WithValue(ctx, contextKeyServer, s)

		// This is a little tricky, as req.Request and Context refer to each other
		req.Request = req.Request.WithContext(ctx)

		response, err := next(ctx, req)

		if err != nil {
			if response != nil {
				// Can return a response to override the default error
				klog.Warningf("error serving request: %v", err)
				response.WriteTo(ctx, w)
				return
			} else {
				klog.Warningf("internal error serving request: %v", err)
				http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
				return
			}
		}

		response.WriteTo(ctx, w)
	}
}

func (s *Server) NewScope(ctx context.Context) *scopes.Scope {
	data := scopes.NewScope()

	for _, component := range s.Components {
		component.AddToScope(ctx, data)
	}

	return data
}
