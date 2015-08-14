package app

import (
	"fmt"
	clog "github.com/drichardson/go-appengine-vm-example/contextlog"
	"github.com/drichardson/go-appengine-vm-example/handler"
	"golang.org/x/net/context"
	"golang.org/x/oauth2/google"
	storage "google.golang.org/api/storage/v1"
	"google.golang.org/appengine/datastore"
	"io"
	"io/ioutil"
	"net/http"
	"strconv"
	"strings"
	"time"
)

func RegisterHandlers() {
	// Simple Request/Response with context.Context
	handler.Handle("/ping", handlePing)

	// App Engine Datastore
	handler.Handle("/datastore/put", handleDatastorePut)
	handler.Handle("/datastore/get", handleDatastoreGet)

	// Google Cloud Storage
	handler.Handle("/storage/put", handleStoragePut)
	handler.Handle("/storage/get", handleStorageGet)

	// Using context.Context to control top level request timeouts
	// while invoking multiple sub-requests.
	handler.Handle("/slow/get", handleSlowGet)
	handler.Handle("/subrequests/serial", handleSerialSubrequests)
	handler.Handle("/subrequests/concurrent", handleConcurrentSubrequests)
}

func handlePing(c context.Context, w http.ResponseWriter, r *http.Request) {
	w.Write([]byte("ping!"))
}

const ExampleKind = "AppEngineManagedVMExample"

type ExampleData struct {
	StringValue string
	IntValue    int
}

func handleDatastorePut(c context.Context, w http.ResponseWriter, r *http.Request) {
	k := r.URL.Query().Get("key")
	s := r.URL.Query().Get("stringvalue")
	iStr := r.URL.Query().Get("intvalue")
	if k == "" || s == "" || iStr == "" {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte("Missing required query parameter key, stringvalue, or intvalue"))
		return
	}
	i, err := strconv.ParseInt(iStr, 10, 64)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte("Couldn't convert intvalue to integer. " + err.Error()))
		return
	}
	key := datastore.NewKey(c, ExampleKind, k, 0, nil)
	if key.Incomplete() {
		// This shouldn't happen because we already made sure k != "", but
		// if it does we want to know because we won't be able to retreive
		// the results with the same key.
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte("Unexpected incomplete key"))
		return
	}
	_, err = datastore.Put(c, key, &ExampleData{StringValue: s, IntValue: int(i)})
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte("Failed to put value to datastore. " + err.Error()))
		return
	}
	w.Write([]byte("ok"))
}

func handleDatastoreGet(c context.Context, w http.ResponseWriter, r *http.Request) {
	k := r.URL.Query().Get("key")
	if k == "" {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte("Missing required query parameter key"))
		return
	}
	key := datastore.NewKey(c, ExampleKind, k, 0, nil)
	if key.Incomplete() {
		// This shouldn't happen because we already made sure k != "", but
		// if it does we want to know because we won't be able to retreive
		// the results with the same key.
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte("Unexpected incomplete key"))
		return
	}
	e := new(ExampleData)
	err := datastore.Get(c, key, &e)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte("Datastore get failed. " + err.Error()))
		return
	}

	w.Write([]byte(fmt.Sprintf("StringValue: %v, IntValue: %v", e.StringValue, e.IntValue)))
}

func handleStoragePut(c context.Context, w http.ResponseWriter, r *http.Request) {
	bucket := r.URL.Query().Get("bucket")
	name := r.URL.Query().Get("name")
	value := r.URL.Query().Get("value")
	if bucket == "" || name == "" || value == "" {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte("Missing bucket, name, or value query parameter."))
		return
	}
	client, err := google.DefaultClient(c, storage.DevstorageReadWriteScope)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte("Failed to get default google client. " + err.Error()))
		return
	}
	service, err := storage.New(client)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte("Failed to get storage service. " + err.Error()))
		return
	}
	obj, err := service.Objects.Insert(bucket, &storage.Object{Name: name}).Media(strings.NewReader(value)).Do()
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte("Failed to insert object. " + err.Error()))
		return
	}
	w.Write([]byte(fmt.Sprintf("put succeeded: %v", obj)))
}

func handleStorageGet(c context.Context, w http.ResponseWriter, r *http.Request) {
	bucket := r.URL.Query().Get("bucket")
	name := r.URL.Query().Get("name")
	if bucket == "" || name == "" {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte("Missing bucket or name query parameter."))
		return
	}
	client, err := google.DefaultClient(c, storage.DevstorageReadOnlyScope)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte("Failed to get default google client. " + err.Error()))
		return
	}
	service, err := storage.New(client)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte("Failed to get storage service. " + err.Error()))
		return
	}
	res, err := service.Objects.Get(bucket, name).Download()
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte("Failed to get object. " + err.Error()))
		return
	}
	w.Header().Set("Content-Type", res.Header.Get("Content-Type"))
	_, err = io.Copy(w, res.Body)
	if err != nil {
		// to late to change status code now
		clog.Errorf(c, "io.Copy failed to copy storage get to response. %v", err)
	}
}

func handleSlowGet(c context.Context, w http.ResponseWriter, r *http.Request) {
	delayStr := r.URL.Query().Get("delay")
	if delayStr == "" {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte("missing delay (e.g., delay=300ms) query parameter"))
		return
	}
	delay, err := time.ParseDuration(delayStr)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte("invalid delay duration string. " + err.Error()))
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
		w.Write([]byte("Invalid timeout query parameter. Expected something like timeout=700ms"))
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
			w.Write([]byte(fmt.Sprintf("Sub-request %v failed. %v", i, err.Error())))
			return
		}
		subResults = append(subResults, r)
	}

	duration := time.Since(start)

	result := fmt.Sprintf("Duration: %v\nResults:%v", duration, strings.Join(subResults, "\n"))
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
		w.Write([]byte("Invalid timeout query parameter. Expected something like timeout=700ms"))
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
			w.Write([]byte(fmt.Sprintf("Sub request %v failed. %v", i, err.Error())))
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

	result := fmt.Sprintf("Duration: %v\n%v\n%v", duration, toString(r1), toString(r2))
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
