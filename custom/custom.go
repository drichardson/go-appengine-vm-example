// This example only works on Managed VMs. Don't use goapp (the
// classic environment for Go App Engine) to build.

package main

import (
	"github.com/drichardson/go-appengine-vm-example/app"
	"log"
	"net/http"
)

func main() {
	app.RegisterHandlers()
	err := http.ListenAndServe(":8080", nil)
	log.Fatal(err)
}
