package app

import (
	"fmt"
	clog "github.com/drichardson/go-appengine-vm-example/contextlog"
	"github.com/drichardson/go-appengine-vm-example/handler"
	"golang.org/x/net/context"
	"io/ioutil"
	"net/http"
	"time"
)

func RegisterHandlers() {
	handler.Handle("/ping", handlePing)
	handler.Handle("/datastore/put", handleDatastorePut)
	handler.Handle("/datastore/get", handleDatastoreGet)
	handler.Handle("/slow/get", handleSlowGet)
	handler.Handle("/subrequests/serial", handleSerialSubrequests)
	handler.Handle("/subrequests/concurrent", handleConcurrentSubrequests)
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
	w.Write([]byte(fmt.Sprintf("delayed %v\n", delayStr)))
}

func localURL(path string) string {
	return "http://localhost:8080/" + path
}

func handleSerialSubrequests(ctx context.Context, w http.ResponseWriter, r *http.Request) {

	// Require the caller to specify an overall timeout, like 700ms.
	timeoutStr := r.URL.Query().Get("timeout")
	timeout, err := time.ParseDuration(timeoutStr)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte("Invalid timeout query parameter. Expected something like timeout=700ms\n"))
		return
	}

	ctx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	var sub1Result, sub2Result string

	start := time.Now()

	for i := 0; i < 2; i++ {
		err = httpGet(ctx, localURL("slow/get?delay=500ms"), func(r *http.Response, err error) error {
			if err != nil {
				clog.Debugf(ctx, "handleTimeout: get %v failed.", i, err)
				return err
			}
			b, err := ioutil.ReadAll(r.Body)
			if err != nil {
				clog.Debug(ctx, "handleTimeout: get %v ReadAll failed.", i, err)
				return err
			}
			clog.Debug(ctx, "handleTimeout: get %v returned with:", i, string(b))
			sub1Result = string(b)
			return nil
		})

		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte(fmt.Sprintf("Sub request %v failed. %v\n", i, err.Error())))
			return
		}
	}

	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte("Second sub request failed. " + err.Error() + "\n"))
		return
	}

	duration := time.Since(start)

	result := fmt.Sprintf("Duration: %v\nFirst: %vSecond: %v", duration, sub1Result, sub2Result)
	w.Write([]byte(result))
}

func handleConcurrentSubrequests(ctx context.Context, w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusNotImplemented)
}

func httpGet(ctx context.Context, url string, f func(*http.Response, error) error) error {
	r, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return err
	}
	return httpDo(ctx, r, f)
}

// Taken from (Sameer Ajmani's blog post)[https://blog.golang.org/context].
// Runs the HTTP request and processes its response in a new goroutine. It
// cancels the request if ctx.Done is closed before the goroutine exits:
func httpDo(ctx context.Context, req *http.Request, f func(*http.Response, error) error) error {
	// Run the HTTP request in a goroutine and pass the response to f.
	tr := &http.Transport{}
	client := &http.Client{Transport: tr}
	c := make(chan error, 1)
	go func() { c <- f(client.Do(req)) }()
	select {
	case <-ctx.Done():
		tr.CancelRequest(req)
		<-c // Wait for f to return.
		return ctx.Err()
	case err := <-c:
		return err
	}
}
