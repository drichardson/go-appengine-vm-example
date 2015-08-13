package handler

import (
	"golang.org/x/net/context"
	"net/http"
)

type ContextHandler func(context.Context, http.ResponseWriter, *http.Request)

func Handle(pattern string, handler ContextHandler) {
	http.HandleFunc(pattern, func(w http.ResponseWriter, r *http.Request) {
		// contextWrapper is envrionment dependent, so it's in its own file.
		handler(contextWrapper(r), w, r)
	})
}
