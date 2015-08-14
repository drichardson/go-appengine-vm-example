// An App Engine Managed VM Custom Runtime, as described here:
// https://cloud.google.com/appengine/docs/managed-vms/custom-runtimes

package main

import (
	"github.com/drichardson/go-appengine-vm-example/app"
	clog "github.com/drichardson/go-appengine-vm-example/contextlog"
	"github.com/drichardson/go-appengine-vm-example/handler"
	"golang.org/x/net/context"
	"log"
	"net/http"
)

func main() {
	app.RegisterHandlers()
	registerLifecycleEventHandlers()
	err := http.ListenAndServe(":8080", nil)
	log.Fatal(err)
}

// App Engine Managed VMs Custom Runtimes must implement the App Engine
// lifecycle handlers.
// https://cloud.google.com/appengine/docs/managed-vms/custom-runtimes
func registerLifecycleEventHandlers() {
	handler.Handle("/_ah/start", okHandler)
	handler.Handle("/_ah/stop", okHandler)
	handler.Handle("/_ah/health", okHandler)
}

func okHandler(c context.Context, w http.ResponseWriter, r *http.Request) {
	clog.Debugf(c, "Responding with 'ok' for %v", r.URL)
	w.Write([]byte("ok"))
}
