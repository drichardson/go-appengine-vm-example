// +build !appengine

package handler

import (
	"golang.org/x/net/context"
	"net/http"
)

func contextWrapper(r *http.Request) context.Context {
	return context.Background()
}
