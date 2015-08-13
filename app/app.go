package app

import (
	clog "github.com/drichardson/go-appengine-vm-example/contextlog"
	"github.com/drichardson/go-appengine-vm-example/handler"
	"golang.org/x/net/context"
	"net/http"
	"time"
)

func RegisterHandlers() {
	handler.Handle("/ping", handlePing)
	handler.Handle("/datastore/put", handleDatastorePut)
	handler.Handle("/datastore/get", handleDatastoreGet)
	handler.Handle("/slow/get", handleSlowGet)
}

func handlePing(c context.Context, w http.ResponseWriter, r *http.Request) {
	clog.Debug(c, "handlePing called")
	w.Write([]byte("ok\n"))
}

func handleDatastorePut(c context.Context, w http.ResponseWriter, r *http.Request) {
	clog.Debug(c, "handleDatastorePut called")
	w.Write([]byte("ok\n"))
}

func handleDatastoreGet(c context.Context, w http.ResponseWriter, r *http.Request) {
	clog.Debug(c, "handleDatastoreGet called")
	w.Write([]byte("ok\n"))
}

func handleSlowGet(c context.Context, w http.ResponseWriter, r *http.Request) {
	delayStr := r.URL.Query().Get("delay")
	if delayStr == "" {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte("missing delay (e.g., delay=300ms) query parameter\n"))
		return
	}
	delay, err := time.ParseDuration(delayStr)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte("invalid delay duration string. " + err.Error() + "\n"))
		return
	}
	time.Sleep(delay)
	w.WriteHeader(http.StatusNoContent)
}

func handleTimeout(ctx context.Context, w http.ResponseWriter, r *http.Request) {
	/*
		ctx, cancel := context.WithTimeout(ctx, time.Duration(1)*time.Second)
		defer cancel()

		c := make(chan error, 1)
		go func() {
			clog.Debug("Sleeping 5 seconds")
			time.Sleep(time.Duration(5) * time.Second)
			clog.Debug("Done sleeping")
		}()

		select {
		case <-ctx.Done():
		}
	*/
}
