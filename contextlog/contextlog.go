// +build !appengine

package contextlog

import (
	"fmt"
	"golang.org/x/net/context"
	"log"
)

func Debug(c context.Context, args ...interface{}) {
	log.Println("DEBUG: " + fmt.Sprint(args...))
}

func Debugf(c context.Context, format string, args ...interface{}) {
	log.Printf("DEBUG: "+format, args...)
}

func Error(c context.Context, args ...interface{}) {
	log.Println("ERROR: " + fmt.Sprint(args...))
}

func Errorf(c context.Context, format string, args ...interface{}) {
	log.Printf("ERROR: "+format, args...)
}
