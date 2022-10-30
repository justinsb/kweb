package components

import (
	"net"
	"strings"

	"k8s.io/klog/v2"
)

func (r *Request) IsLocalhost() bool {
	host := r.Request.Host
	if !strings.HasPrefix(host, "localhost") {
		return false
	}
	if host == "localhost" {
		return true
	}
	if strings.HasPrefix(host, "localhost:") {
		h, _, err := net.SplitHostPort(host)
		if err == nil && h == "localhost" {
			return true
		}
	}
	return false
}

// BrowserUsingHTTPS returns true if the _browser_ is using https to talk to us
// We may (or may not) be using encryption behind a load balancer
func (r *Request) BrowserUsingHTTPS() bool {
	forwardedProto := r.Header.Get("X-Forwarded-Proto")
	switch forwardedProto {
	case "":
	case "http":
		return false
	case "https":
		return true
	default:
		klog.Warningf("unknown x-forwarded-proto header %q", forwardedProto)
	}

	if r.URL.Scheme == "https" {
		return true
	}

	return false
}
