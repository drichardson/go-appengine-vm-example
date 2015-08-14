// This example only works on Managed VMs. Don't use goapp (the
// classic environment for Go App Engine) to build.

package main

import (
	"github.com/drichardson/go-appengine-vm-example/app"
	"google.golang.org/appengine"
)

func main() {
	app.RegisterHandlers()
	appengine.Main()
}
