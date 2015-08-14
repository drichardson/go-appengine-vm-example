package app

import (
	"fmt"
	clog "github.com/drichardson/go-appengine-vm-example/contextlog"
	"github.com/drichardson/go-appengine-vm-example/handler"
	"golang.org/x/net/context"
	"io/ioutil"
	"net/http"
	"strings"
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
	w.Write([]byte(fmt.Sprintf("delayed %v", delayStr)))
}

func localURL(path string) string {
	return "http://localhost:8080/" + path
}

// handleSerialSubrequests executes two slow running sub-requests
// serially with a timeout.
//
// both sub-requests succeed:
//	curl localhost:8080/subrequests/serial?timeout=1100ms
// one fails, one succeeds:
//	curl localhost:8080/subrequests/serial?timeout=700ms
// both fail:
//	curl localhost:8080/subrequests/serial?timeout=400ms
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

	start := time.Now()

	get := func(i int) (string, error) {
		var result string
		err := httpGet(ctx, localURL("slow/get?delay=500ms"), func(r *http.Response, err error) error {
			if err != nil {
				clog.Debugf(ctx, "handleTimeout: get %v failed. %v", i, err)
				return err
			}
			b, err := ioutil.ReadAll(r.Body)
			if err != nil {
				clog.Debugf(ctx, "handleTimeout: get %v ReadAll failed. %v", i, err)
				return err
			}
			clog.Debugf(ctx, "handleTimeout: get %v returned with: %v", i, string(b))
			result = string(b)
			return nil
		})
		return result, err
	}

	subRequestCount := 2
	subResults := make([]string, 0)
	for i := 0; i < subRequestCount; i++ {
		r, err := get(i)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte(fmt.Sprintf("Sub-request %v failed. %v\n", i, err.Error())))
			return
		}
		subResults = append(subResults, r)
	}

	duration := time.Since(start)

	result := fmt.Sprintf("Duration: %v\nResults:\n%v\n", duration, strings.Join(subResults, "\n"))
	w.Write([]byte(result))
}

// handleConhandleConcurrentSubrequests executes two slow running sub-requests
// concurrently with a timeout. If the timeout expires, the results it has
// obtained so far are returned.
//
// both sub-requests succeed:
//	curl localhost:8080/subrequests/concurrent?timeout=800ms
// one fails, one succeeds:
//	curl localhost:8080/subrequests/concurrent?timeout=700ms
// both fail:
//	curl localhost:8080/subrequests/concurrent?timeout=400ms
func handleConcurrentSubrequests(ctx context.Context, w http.ResponseWriter, r *http.Request) {
	// Require the caller to specify an overall timeout, like 700ms.
	clog.Debug(ctx, "In handleConcurrentSubrequests test log message")
	timeoutStr := r.URL.Query().Get("timeout")
	timeout, err := time.ParseDuration(timeoutStr)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte("Invalid timeout query parameter. Expected something like timeout=700ms\n"))
		return
	}

	ctx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	type subRequestResult struct {
		err           error
		requestNumber int
		result        string
	}

	out := make(chan subRequestResult)

	start := time.Now()

	subRequestDelays := []time.Duration{500 * time.Millisecond, 750 * time.Millisecond}
	for i, delay := range subRequestDelays {
		i := i         // shadow mutable counter for closure
		delay := delay // shadow for closure
		go func() {
			var result string
			url := localURL(fmt.Sprintf("slow/get?delay=%v", delay))
			err := httpGet(ctx, url, func(r *http.Response, err error) error {
				if err != nil {
					clog.Debugf(ctx, "handleTimeout: get %v failed. %v", i, err)
					return err
				}
				b, err := ioutil.ReadAll(r.Body)
				if err != nil {
					clog.Debugf(ctx, "handleTimeout: get %v ReadAll failed. %v", i, err)
					return err
				}
				clog.Debugf(ctx, "handleTimeout: get %v returned with: %v", i, string(b))
				result = string(b)
				return nil
			})
			out <- subRequestResult{err, i, result}
		}()

		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte(fmt.Sprintf("Sub request %v failed. %v\n", i, err.Error())))
			return
		}
	}

	r1 := <-out
	r2 := <-out

	duration := time.Since(start)

	toString := func(s subRequestResult) string {
		if s.err != nil {
			return fmt.Sprintf("Request Number: %v, Error: %v", s.requestNumber, s.err.Error())
		} else {
			return fmt.Sprintf("Request Number: %v, Result: %v", s.requestNumber, s.result)
		}
	}

	result := fmt.Sprintf("Duration: %v\n%v\n%v\n", duration, toString(r1), toString(r2))
	w.Write([]byte(result))
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
