package server

import (
	"fmt"
	"net/http"

	"k8s.io/klog/v2"
)

func (s *Server) ensureMux() error {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	if s.mux != nil {
		return nil
	}

	mux := http.NewServeMux()
	for _, component := range s.Components {
		if err := component.RegisterHandlers(&s.Server, mux); err != nil {
			return fmt.Errorf("error registering component handlers for %T: %w", component, err)
		}
	}

	s.mux = mux
	return nil
}

func (s *Server) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	klog.Infof("%s %s", r.Method, r.URL)

	s.mux.ServeHTTP(w, r)
}

type ErrorHandler struct {
	Status int
}

func (m *ErrorHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	klog.Infof("ErrorHandler %v: %+v", m.Status, r)
	http.Error(w, http.StatusText(m.Status), m.Status)
	return
}
