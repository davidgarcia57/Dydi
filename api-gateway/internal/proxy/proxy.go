package proxy

import (
	"errors"
	"net"
	"net/http"
	"net/http/httputil"
	"net/url"
	"strings"
	"time"
)

// proxyTransport bounds how long the gateway waits on a downstream service.
// ResponseHeaderTimeout is generous (60s) so a cold-starting Render service can
// still answer the first request instead of being cut off, while a truly hung
// upstream still fails instead of tying up the connection forever.
var proxyTransport = &http.Transport{
	Proxy: http.ProxyFromEnvironment,
	DialContext: (&net.Dialer{
		Timeout:   10 * time.Second,
		KeepAlive: 30 * time.Second,
	}).DialContext,
	ForceAttemptHTTP2:     true,
	MaxIdleConns:          100,
	IdleConnTimeout:       90 * time.Second,
	TLSHandshakeTimeout:   10 * time.Second,
	ExpectContinueTimeout: 1 * time.Second,
	ResponseHeaderTimeout: 60 * time.Second,
}

// errorHandler returns a clean JSON 502/504 instead of the default plain-text
// "502 Bad Gateway" with a Go error string. The frontend's api.js retries on
// 502-504, which pairs with Render cold starts.
func errorHandler(w http.ResponseWriter, _ *http.Request, err error) {
	status := http.StatusBadGateway
	var nerr net.Error
	if errors.As(err, &nerr) && nerr.Timeout() {
		status = http.StatusGatewayTimeout
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	w.Write([]byte(`{"error":"upstream service unavailable"}`)) //nolint:errcheck
}

func To(target string) http.Handler {
	u, err := url.Parse(target)
	if err != nil {
		panic("invalid proxy target: " + target)
	}
	rp := httputil.NewSingleHostReverseProxy(u)
	rp.Transport = proxyTransport
	rp.ErrorHandler = errorHandler
	rp.Director = func(r *http.Request) {
		r.URL.Scheme = u.Scheme
		r.URL.Host = u.Host
		r.Host = u.Host
		r.URL.Path = downstreamPath(u.Path, r.URL.Path)
	}
	return rp
}

func downstreamPath(targetPath, requestPath string) string {
	path := strings.TrimPrefix(requestPath, "/api")
	if targetPath != "" && targetPath != "/" {
		path = strings.TrimRight(targetPath, "/") + "/" + strings.TrimLeft(path, "/")
	}
	if path == "" {
		return "/"
	}
	if !strings.HasPrefix(path, "/") {
		return "/" + path
	}
	return path
}

func WebSocket(target string) http.Handler {
	u, err := url.Parse(target)
	if err != nil {
		panic("invalid ws proxy target: " + target)
	}
	rp := httputil.NewSingleHostReverseProxy(u)
	rp.ErrorHandler = errorHandler
	rp.Director = func(r *http.Request) {
		r.URL.Scheme = u.Scheme
		r.URL.Host = u.Host
		r.Host = u.Host
	}
	return rp
}
