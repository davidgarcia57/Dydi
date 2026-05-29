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
		// strip the /api prefix that chi already consumed
		if i := strings.Index(r.URL.Path, u.Path); i >= 0 {
			r.URL.Path = r.URL.Path[i:]
		}
	}
	return rp
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
