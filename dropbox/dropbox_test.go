// Copyright (c) 2015, Álvaro Vilanova Vidal
// Use of this source code is governed by a BSD 3-Clause license that can be
// found in the LICENSE file.

package dropbox

import (
	"bytes"
	"errors"
	"io"
	"io/ioutil"
	"net/http"
	"testing/iotest"

	"testing"
)

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
	if got, want := checkResponse(res), error(nil); got != want {
		t.Errorf("checkResponse({StatusCode: 200}) returned %v, want %v", got, want)
	}
	res.StatusCode = 300
	if got, want := checkResponse(res), error(UnexpectedError{}); got != want {
		t.Errorf("checkResponse({StatusCode: 300}) returned %v, want %v", got, want)
	}
	res.StatusCode = 500
	if got, want := checkResponse(res), error(UnexpectedError{}); got != want {
		t.Errorf("checkResponse({StatusCode: 500}) returned %v, want %v", got, want)
	}
	res.Body = ioutil.NopCloser(bytes.NewBufferString("db-error"))
	if got, want := checkResponse(res), error(UnexpectedError{}); got != want {
		t.Errorf("checkResponse({StatusCode: 500, No Content-Type, Body:\"db-error\"}) returned %v, want %v", got, want)
	}
	res.Header = http.Header{}
	res.Header.Add("Content-Type", "text/plain")
	res.Body = ioutil.NopCloser(bytes.NewBufferString("db-error"))
	if got, want := checkResponse(res), error(Error{"db-error"}); got != want {
		t.Errorf("checkResponse({StatusCode: 500, \"text/plain\", Body:\"db-error\"}) returned %v, want %v", got, want)
	}
	res.Body = ioutil.NopCloser(iotest.TimeoutReader(bytes.NewBufferString("db-error")))
	if got, want := checkResponse(res), error(Error{"db-error"}); got == want {
		t.Errorf("checkResponse with an invalid text content must return an io error")
	}
	res.Header.Set("Content-Type", "application/json")
	res.Body = ioutil.NopCloser(bytes.NewBufferString("db-error"))
	if got, want := checkResponse(res), error(Error{"db-error"}); got == want {
		t.Errorf("checkResponse with an invalid JSON body must return an encoding error")
	}
	res.Body = ioutil.NopCloser(bytes.NewBufferString("{\"reason\":\"db-error\"}"))
	if got, want := checkResponse(res), error(Error{"db-error"}); got != want {
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
		t.Errorf("newRequest(invalid URL) returned no error")
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
		t.Errorf("newRequest(...) expected no error, got %v", err)
	}
	blob, _ := ioutil.ReadAll(req.Body)
	bodyStr := bytes.NewBuffer(blob).String()
	if got, want := bodyStr, "a nice body"; got != want {
		t.Errorf("newRequest(...) Request.Body is %v, want %v", got, want)
	}
}
