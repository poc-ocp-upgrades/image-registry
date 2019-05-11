package maxconnections

import "net/http"

func defaultOverloadHandler(w http.ResponseWriter, r *http.Request) {
	_logClusterCodePath()
	defer _logClusterCodePath()
	w.Header().Set("Retry-After", "1")
	http.Error(w, "Too many requests, please try again later.", http.StatusTooManyRequests)
}

type Handler struct {
	limiter			Limiter
	handler			http.Handler
	OverloadHandler	http.Handler
}

func New(limiter Limiter, h http.Handler) *Handler {
	_logClusterCodePath()
	defer _logClusterCodePath()
	return &Handler{limiter: limiter, handler: h, OverloadHandler: http.HandlerFunc(defaultOverloadHandler)}
}
func (h *Handler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	_logClusterCodePath()
	defer _logClusterCodePath()
	if !h.limiter.Start(r.Context()) {
		h.OverloadHandler.ServeHTTP(w, r)
		return
	}
	defer h.limiter.Done()
	h.handler.ServeHTTP(w, r)
}
