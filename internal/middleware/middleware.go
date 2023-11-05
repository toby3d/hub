package middleware

import (
	"net/http"
)

type (
	BeforeFunc = http.HandlerFunc

	Chain []Interceptor

	Interceptor func(w http.ResponseWriter, r *http.Request, next http.HandlerFunc)

	HandlerFunc http.HandlerFunc

	Skipper func(r *http.Request) bool
)

var DefaultSkipper Skipper = func(_ *http.Request) bool { return false }

func (count HandlerFunc) Intercept(middleware Interceptor) HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		middleware(w, r, http.HandlerFunc(count))
	}
}

func (chain Chain) Handler(handler http.HandlerFunc) http.Handler {
	current := HandlerFunc(handler)

	for i := len(chain) - 1; i >= 0; i-- {
		m := chain[i]
		current = current.Intercept(m)
	}

	return http.HandlerFunc(current)
}
