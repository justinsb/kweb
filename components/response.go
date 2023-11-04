package components

import (
	"context"
	"encoding/json"
	"net/http"

	"k8s.io/klog/v2"
)

type staticResponse struct {
	StatusCode  int
	StatusText  string
	RedirectURL string
}

func (r staticResponse) WriteTo(ctx context.Context, w http.ResponseWriter) {
	statusCode := r.StatusCode

	statusText := r.StatusText
	if statusText == "" {
		statusText = http.StatusText(statusCode)
	}
	if r.RedirectURL != "" {
		httpRequest := GetRequest(ctx)
		http.Redirect(w, httpRequest.Request, r.RedirectURL, r.StatusCode)
	} else {
		http.Error(w, statusText, statusCode)
	}
}

func ErrorResponse(code int) Response {
	r := &staticResponse{}
	return r.WithStatus(code)
}

func (r *staticResponse) WithStatus(code int) *staticResponse {
	r.StatusCode = code
	return r
}

func RedirectResponse(redirectURL string) Response {
	r := &staticResponse{}

	r.StatusCode = http.StatusFound
	r.RedirectURL = redirectURL
	return r
}

type SimpleResponse struct {
	StatusCode int
	StatusText string
	headers    http.Header
	Body       []byte
}

func (r *SimpleResponse) Headers() http.Header {
	if r.headers == nil {
		r.headers = make(http.Header)
	}
	return r.headers
}
func (r SimpleResponse) WriteTo(ctx context.Context, w http.ResponseWriter) {
	statusCode := r.StatusCode
	if statusCode == 0 {
		statusCode = 200
	}
	statusText := r.StatusText
	if statusText == "" {
		statusText = http.StatusText(statusCode)
	}

	for k, values := range r.headers {
		for _, v := range values {
			w.Header().Add(k, v)
		}
	}
	w.WriteHeader(statusCode)
	w.Write(r.Body)
}

type JSONResponse struct {
	Object any
}

func (r JSONResponse) WriteTo(ctx context.Context, w http.ResponseWriter) {
	b, err := json.Marshal(r.Object)
	if err != nil {
		klog.Warningf("error from json.Marshal(%T): %v", r.Object, err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusOK)
	w.Write(b)
}
