package components

import (
	"context"
	"net/http"
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
	Body       []byte
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

	w.WriteHeader(statusCode)
	w.Write(r.Body)
}
