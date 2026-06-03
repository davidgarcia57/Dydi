package proxy

import (
	"net/http"
	"net/http/httputil"
	"net/url"
	"strings"
)

func To(target string) http.Handler {
	u, err := url.Parse(target)
	if err != nil {
		panic("invalid proxy target: " + target)
	}
	rp := httputil.NewSingleHostReverseProxy(u)
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
	rp.Director = func(r *http.Request) {
		r.URL.Scheme = u.Scheme
		r.URL.Host = u.Host
		r.Host = u.Host
	}
	return rp
}
