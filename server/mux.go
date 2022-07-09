package server

import (
	"net/http"

	"k8s.io/klog/v2"
)

func (s *Server) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	klog.Infof("%s %s", r.Method, r.URL)

	mux := http.NewServeMux()
	for _, component := range s.Components {
		component.RegisterHandlers(&s.Server, mux)
	}

	// Fallback
	mux.Handle("/", &ErrorHandler{Status: http.StatusNotFound})

	mux.ServeHTTP(w, r)
}

type ErrorHandler struct {
	Status int
}

func (m *ErrorHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	http.Error(w, http.StatusText(m.Status), m.Status)
	return
}
