package components

import (
	"context"
	"fmt"
	"net/http"

	"k8s.io/klog/v2"
)

type Server struct {
	Components []Component
}

func GetComponent[T any](s *Server, dest *T) error {
	var matches []T
	for _, component := range s.Components {
		match, ok := component.(T)
		if ok {
			matches = append(matches, match)
		}
	}
	if len(matches) == 0 {
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
		ctx = context.WithValue(ctx, contextKeyRequest, req)

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
