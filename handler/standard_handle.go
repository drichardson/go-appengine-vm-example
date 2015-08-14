// +build appenginevm

package handler

import (
	"golang.org/x/net/context"
	"google.golang.org/appengine"
	"net/http"
)

func contextWrapper(r *http.Request) context.Context {
	return appengine.NewContext(r)
}
