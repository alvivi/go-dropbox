// Copyright (c) 2015, Álvaro Vilanova Vidal
// Use of this source code is governed by a BSD 3-Clause license that can be
// found in the LICENSE file.

package dropbox

import (
	"bytes"
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
