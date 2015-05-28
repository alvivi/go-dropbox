// Copyright (c) 2015, Álvaro Vilanova Vidal
// Use of this source code is governed by a BSD 3-Clause license that can be
// found in the LICENSE file.

package dropbox

import (
	"net/http"

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
