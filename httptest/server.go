package util

import (
	"fmt"
	"io/ioutil"
	"net"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/davecgh/go-spew/spew"
	"github.com/stretchr/testify/assert"
)

type ServerConfig struct {
	t                  *testing.T
	RequestMethod      string
	RequestURLPath     string
	RequestBody        []byte
	RequestHeaders     map[string]string
	RequestQueryParams map[string]string
	ResponseHeaders    map[string]string
	ResponseBody       []byte
	ResponseBodyFunc   func(t *testing.T, actualBody []byte)
	ResponseStatusCode int
	HookAfterResponse  func()
	Debug              bool
}

func (cfg *ServerConfig) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	ass := assert.New(cfg.t)
	defer r.Body.Close()

	if cfg.RequestMethod != "" {
		if cfg.Debug {
			fmt.Fprintf(os.Stderr, "Method: ")
			spew.Fdump(os.Stderr, r.Method)
		}
		ass.Equal(cfg.RequestMethod, r.Method, "request method mismatch")
	}

	if cfg.RequestURLPath != "" {
		if cfg.Debug {
			fmt.Fprintf(os.Stderr, "Path: ")
			spew.Fdump(os.Stderr, r.URL.Path)
		}
		ass.Equal(cfg.RequestURLPath, r.URL.Path, "request path mismatch")
	}

	if cfg.Debug && len(cfg.RequestHeaders) > 0 {
		if cfg.Debug {
			fmt.Fprintf(os.Stderr, "Header: ")
			spew.Fdump(os.Stderr, r.Header)
		}
	}
	reqHeaders := r.Header
	for k, want := range cfg.RequestHeaders {
		actual := reqHeaders.Get(k)
		ass.Equal(want, actual, fmt.Sprintf("header '%s' mismatch", k))
	}

	if len(cfg.RequestQueryParams) > 0 {
		if cfg.Debug {
			fmt.Fprintf(os.Stderr, "QueryParams: ")
			spew.Fdump(os.Stderr, r.URL.RawQuery)
		}
	}
	reqQueryParams := r.URL.Query()
	for k, want := range cfg.RequestQueryParams {
		actual := reqQueryParams.Get(k)
		ass.Equal(want, actual, fmt.Sprintf("query param '%s' mismatch", k))
	}

	if len(cfg.RequestBody) > 0 {
		body, _ := ioutil.ReadAll(r.Body)
		if cfg.Debug {
			fmt.Fprintf(os.Stderr, "Body: ")
			spew.Fdump(os.Stderr, body)
		}
		if cfg.ResponseBodyFunc != nil {
			cfg.ResponseBodyFunc(cfg.t, body)
		} else {
			ass.Equal(cfg.RequestBody, body, "request body mismatch")
		}
	}

	// must in this order: header -> status code -> body
	rspHeaders := w.Header()
	for k, v := range cfg.ResponseHeaders {
		rspHeaders.Set(k, v)
	}

	if cfg.ResponseStatusCode > 0 {
		w.WriteHeader(cfg.ResponseStatusCode)
	}

	if len(cfg.ResponseBody) > 0 {
		w.Write(cfg.ResponseBody)
	}

	if cfg.HookAfterResponse != nil {
		cfg.HookAfterResponse()
	}
}

func NewServer(t *testing.T, config *ServerConfig) *httptest.Server {
	return httptest.NewServer(config)
}

// NewUnstartedServer returns a new Server but doesn't start it.
//
// After changing its configuration, the caller should call Start or
// StartTLS.
//
// The caller should call Close when finished, to shut it down.
func NewUnstartedServer(t *testing.T, config *ServerConfig) *httptest.Server {
	return httptest.NewUnstartedServer(config)
}

// Start starts a server from NewUnstartedServer. listening on address
func StartTestServer(server *httptest.Server, address string) {
	server.Listener = newLocalListener(address)
	server.Start()
}

func newLocalListener(address string) net.Listener {
	if address != "" {
		l, err := net.Listen("tcp", address)
		if err != nil {
			panic(fmt.Sprintf("httptest: failed to listen on %v: %v", address, err))
		}
		return l
	}
	l, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		if l, err = net.Listen("tcp6", "[::1]:0"); err != nil {
			panic(fmt.Sprintf("httptest: failed to listen on a port: %v", err))
		}
	}
	return l
}