// +build appengine

package contextlog

import (
	"fmt"
	"golang.org/x/net/context"
	"google.golang.org/appengine/log"
)

func Debug(c context.Context, args ...interface{}) {
	log.Debugf(c, "%v", fmt.Sprint(args...))
}

func Error(c context.Context, args ...interface{}) {
	log.Errorf(c, "%v", fmt.Sprint(args...))
}
