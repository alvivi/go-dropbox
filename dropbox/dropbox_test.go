// Copyright (c) 2015, Álvaro Vilanova Vidal
// Use of this source code is governed by a BSD 3-Clause license that can be
// found in the LICENSE file.

package dropbox

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"net/url"
	"reflect"
	"strings"
	"testing/iotest"

	"testing"
)

var (
	// mux is the HTTP request multiplexer used with the test server.
	mux *http.ServeMux

	// client is the Dropbox client being tested.
	client *Client

	// server is a test HTTP server used to provide mock API responses.
	server *httptest.Server
)

func TestNewRPCRequest(t *testing.T) {
	c := NewClient(nil)

	inBody := struct {
		Bar string `json:"bar"`
		Qux int    `json:"qux"`
	}{"dropbar", 14}
	outBody := "{\"bar\":\"dropbar\",\"qux\":14}"

	req, err := c.NewRPCRequest("POST", "/foo", &inBody)
	if err != nil {
		t.Errorf("TestNewRPCRequest returned unexpected error: %v", err)
	}

	body, _ := ioutil.ReadAll(req.Body)
	if got, want := strings.TrimSpace(string(body)), outBody; got != want {
		t.Errorf("TestNewRPCRequest(%q) Body is %v, want %v", inBody, got, want)
	}

	if got, want := req.Header.Get("Accept"), "application/json"; strings.HasPrefix(want, got) {
		t.Errorf("TestNewRPCRequest Request Accept Header is %v, want something like %v", got, want)
	}

	if got, want := req.Header.Get("Content-Type"), "application/json"; strings.HasPrefix(want, got) {
		t.Errorf("TestNewRPCRequest Request Content-Type Header is %v, want something like %v", got, want)
	}
}

func TestNewRPCRequest_invalidJSON(t *testing.T) {
	c := NewClient(nil)

	type T struct {
		A map[int]interface{}
	}
	_, err := c.NewRPCRequest("GET", "/", &T{})

	if err == nil {
		t.Error("TestNewRPCRequest expected error to be returned.")
	}
	if err, ok := err.(*json.UnsupportedTypeError); !ok {
		t.Errorf("TestNewRPCRequest expected a JSON error; got %#v.", err)
	}
}

func TestDoRPC(t *testing.T) {
	setup()
	defer teardown()

	type foo struct {
		A string
	}

	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		if m := "GET"; m != r.Method {
			t.Errorf("Request method = %v, want %v", r.Method, m)
		}

		defer r.Body.Close()
		body := new(foo)
		err := json.NewDecoder(r.Body).Decode(body)
		if err != nil {
			t.Errorf("TestDoRPC decoding a RPC request body returned unexpected error: %#v", err)
		}
		want := &foo{"in"}
		if !reflect.DeepEqual(body, want) {
			t.Errorf("Response body = %v, want %v", body, want)
		}

		fmt.Fprint(w, `{"A":"out"}`)
	})

	req, _ := client.NewRPCRequest("GET", "/", &foo{"in"})
	body := new(foo)
	client.DoRPC(req, body)

	want := &foo{"out"}
	if !reflect.DeepEqual(body, want) {
		t.Errorf("Response body = %v, want %v", body, want)
	}
}

func TestDoRPC_httpError(t *testing.T) {
	setup()
	defer teardown()

	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "none")
		w.WriteHeader(500)
	})

	req, _ := client.NewRPCRequest("GET", "/", nil)
	_, err := client.DoRPC(req, nil)

	if _, ok := err.(*UnexpectedError); !ok {
		t.Errorf("Expected an UnexpectedError; got %#v.", err)
	}
}

func TestDoRPC_redirectLoop(t *testing.T) {
	setup()
	defer teardown()

	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		http.Redirect(w, r, "/", http.StatusFound)
	})

	req, _ := client.NewRPCRequest("GET", "/", nil)
	_, err := client.DoRPC(req, nil)

	if err == nil {
		t.Error("Expected error to be returned.")
	}
	if err, ok := err.(*url.Error); !ok {
		t.Errorf("Expected a URL error; got %#v.", err)
	}
}

func TestNewClient(t *testing.T) {
	c := NewClient(nil)

	if got, want := c.client, http.DefaultClient; got != want {
		t.Error("NewClient client is not http.DefaultClient")
	}
	if got, want := c.BaseURL.String(), defaultBaseURL; got != want {
		t.Errorf("NewClient BaseURL is %v, want %v", got, want)
	}
	if got, want := c.ContentURL.String(), defaultContentURL; got != want {
		t.Errorf("NewClient ContentURL is %v, want %v", got, want)
	}
	if got, want := c.UserAgent, userAgent; got != want {
		t.Errorf("NewClient UserAgent is %v, want %v", got, want)
	}
}

func TestCheckContentType(t *testing.T) {
	res := &http.Response{}
	if checkContentType(res, "") {
		t.Error("checkContentType on an empty response must return always false")
	}
	res.Header = http.Header{}
	res.Header.Add("Content-Type", "")
	if !checkContentType(res, "") {
		t.Error("checkContentType({Content-Type: \"\"}, \"\") returned false")
	}
	res.Header.Set("Content-Type", "  \n\t   application/json")
	if !checkContentType(res, "") {
		t.Error("checkContentType is no trimming strings")
	}
	res.Header.Set("Content-Type", "application/json;q=0.8")
	if !checkContentType(res, "application/json") {
		t.Error("checkContentType({Content-Type: \"application/json\"}, \"application/json\") returned false")
	}
}

func TestCheckResponse(t *testing.T) {
	res := &http.Response{}
	res.StatusCode = 200
	err := checkResponse(res)
	if err != nil {
		t.Errorf("checkResponse({StatusCode: 200}) expected no error; got %#v.", err)
	}

	res = &http.Response{}
	res.StatusCode = 300
	err = checkResponse(res)
	if err, ok := err.(*UnexpectedError); !ok {
		t.Errorf("checkResponse({StatusCode: 300}) expected an UnexpectedError; got %#v.", err)
	}

	res = &http.Response{}
	res.StatusCode = 500
	err = checkResponse(res)
	if err, ok := err.(*UnexpectedError); !ok {
		t.Errorf("checkResponse({StatusCode: 500}) expected an UnexpectedError; got %#v.", err)
	}

	res = &http.Response{}
	res.StatusCode = 500
	res.Body = ioutil.NopCloser(bytes.NewBufferString("db-error"))
	if got, want := checkResponse(res), &(UnexpectedError{}); !reflect.DeepEqual(got, want) {
		t.Errorf("checkResponse({StatusCode: 500, No Content-Type, Body:\"db-error\"}) returned %v, want %v", got, want)
	}

	res = &http.Response{}
	res.StatusCode = 500
	res.Header = http.Header{}
	res.Header.Add("Content-Type", "text/plain")
	res.Body = ioutil.NopCloser(bytes.NewBufferString("db-error"))
	if got, want := checkResponse(res), &(Error{"db-error"}); !reflect.DeepEqual(got, want) {
		t.Errorf("checkResponse({StatusCode: 500, \"text/plain\", Body:\"db-error\"}) returned %#v, want %#v", got, want)
	}

	res = &http.Response{}
	res.StatusCode = 500
	res.Header = http.Header{}
	res.Header.Add("Content-Type", "text/plain")
	res.Body = ioutil.NopCloser(iotest.TimeoutReader(bytes.NewBufferString("db-error")))
	if err := checkResponse(res); err == nil {
		t.Error("checkResponse({Timout error}) expected error to be returned.")
	}

	res = &http.Response{}
	res.StatusCode = 500
	res.Header = http.Header{}
	res.Header.Set("Content-Type", "application/json")
	res.Body = ioutil.NopCloser(bytes.NewBufferString("db-error"))
	if err, ok := checkResponse(res).(*json.SyntaxError); !ok {
		t.Errorf("checkResponse({StatusCode: 500, \"application/json\", Body:<invalid json>}) expected a *json.SyntaxError error, got %#v", err)
	}

	res = &http.Response{}
	res.StatusCode = 500
	res.Header = http.Header{}
	res.Header.Set("Content-Type", "application/json")
	res.Body = ioutil.NopCloser(bytes.NewBufferString("{\"reason\":\"db-error\"}"))
	if got, want := checkResponse(res), &(Error{"db-error"}); !reflect.DeepEqual(got, want) {
		t.Errorf("checkResponse({StatusCode: 500, \"application/json\", Body:\"{\"reason\":\"db-error\"}\"}) returned %v, want %v", got, want)
	}
}

func TestNewRequest(t *testing.T) {
	c := NewClient(nil)

	inURL, outURL := "/foo", defaultBaseURL+"foo"
	req, err := c.newRequest("GET", inURL, nil)

	if got, want := err, error(nil); got != want {
		t.Errorf("newRequest returned %v, want %v", got, want)
	}

	if got, want := req.URL.String(), outURL; got != want {
		t.Errorf("newRequest(%q) URL is %v, want %v", inURL, got, want)
	}

	if got, want := req.Header.Get("User-Agent"), c.UserAgent; got != want {
		t.Errorf("NewRequest() User-Agent is %v, want %v", got, want)
	}
}

func TestNewRequest_invalidURL(t *testing.T) {
	c := NewClient(nil)
	_, err := c.newRequest("GET", "%$&!", nil)
	if err == nil {
		t.Errorf("Expected error to be returned")
	}
}

func TestNewRequest_bodyWritter(t *testing.T) {
	c := NewClient(nil)
	errorBw := func(w io.Writer) error {
		return errors.New("bw-error")
	}
	_, err := c.newRequest("GET", "/bar", errorBw)
	if got, want := err, errors.New("bw-error"); got.Error() != want.Error() {
		t.Errorf("newRequest(invalid body writer) returned %v, want %v", got, want)
	}

	idBw := func(s string) func(w io.Writer) error {
		return func(w io.Writer) error {
			bytes.NewBufferString(s).WriteTo(w)
			return nil
		}
	}
	req, err := c.newRequest("GET", "/qux", idBw("a nice body"))
	if err != nil {
		t.Errorf("NewRequest returned unexpected error: %#v", err)
	}
	blob, _ := ioutil.ReadAll(req.Body)
	bodyStr := bytes.NewBuffer(blob).String()
	if got, want := bodyStr, "a nice body"; got != want {
		t.Errorf("newRequest(...) Request.Body is %v, want %v", got, want)
	}
}

// Ensure that no User-Agent header is set if the client's UserAgent is empty.
func TestNewRequest_emptyUserAgent(t *testing.T) {
	c := NewClient(nil)
	c.UserAgent = ""
	req, err := c.newRequest("GET", "/", nil)
	if err != nil {
		t.Errorf("NewRequest returned unexpected error: %#v", err)
	}
	if _, ok := req.Header["User-Agent"]; ok {
		t.Error("NewRequest request contains unexpected User-Agent header")
	}
}

func setup() {
	mux = http.NewServeMux()
	server = httptest.NewServer(mux)

	client = NewClient(nil)
	url, _ := url.Parse(server.URL)
	client.BaseURL = url
}

func teardown() {
	server.Close()
}
