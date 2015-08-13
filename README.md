# Go App Engine Managed VM Examples
This repo contains 2 examples of [App Engine Managed VMs](https://cloud.google.com/appengine/docs/managed-vms/):
one using the standard runtime and one using a custom runtime.

Both projects make use of the App Engine Datastore and Google Cloud Storage services.

Both projects also make use of the
[Auto-generated Google APIs for Go](https://github.com/google/google-api-go-client).

## Standard
[Go Standard Managed VM Runtime](https://cloud.google.com/appengine/docs/go/managed-vms/)

The standard runtime VM example is in the `standard/` subdirectory. It uses
the `google.golang.org/appengine` package. It passes control to the standard
library by invoking appengine.Main().

## Custom
[Managed VM Custom Runtimes](https://cloud.google.com/appengine/docs/managed-vms/custom-runtimes)


The custom runtime VM example is in the `custom/` subdirectory. It does NOT use
the `google.golang.org/appengine` package.



