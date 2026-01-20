package proxy

import (
	"net/http"
	"net/http/httputil"
	"net/url"
)

func To(target string) http.Handler {
	u, err := url.Parse(target)
	if err != nil {
		panic("invalid proxy target: " + err.Error())
	}
	proxy := httputil.NewSingleHostReverseProxy(u)
	proxy.Transport = defaultTransport()

	originalDirector := proxy.Director
	proxy.Director = func(req *http.Request) {
		originalDirector(req)
		req.Host = u.Host
	}
	return proxy
}
