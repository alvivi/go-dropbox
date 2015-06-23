// Copyright (c) 2015, Álvaro Vilanova Vidal
// Use of this source code is governed by a BSD 3-Clause license that can be
// found in the LICENSE file.

package dropbox

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"testing"
)

type dropboxPackage struct{}

var (
	dropbox       dropboxPackage
	exampleServer *httptest.Server
	exampleMux    *http.ServeMux
)

func (d dropboxPackage) NewClient(c *http.Client) *Client {
	dc := NewClient(c)
	url, _ := url.Parse(exampleServer.URL)
	dc.BaseURL = url
	return dc
}

func setupExampleServer() {
	exampleMux.HandleFunc("/2-beta/users/get_current_account",
		func(w http.ResponseWriter, r *http.Request) {
			info := AccountInfo{
				Name: Username{
					DisplayName: "Drew",
				},
			}
			json.NewEncoder(w).Encode(info)
		})

	exampleMux.HandleFunc("/2-beta/files/list_folder",
		func(w http.ResponseWriter, r *http.Request) {
			resp := listResponse{
				Entries: []Entry{
					Entry{Name: "James.jpg"},
					Entry{Name: "Mary.jpg"},
					Entry{Name: "Richard.jpg"},
					Entry{Name: "Susan.jpg"},
				},
			}
			json.NewEncoder(w).Encode(resp)
		})
}

func TestMain(m *testing.M) {
	exampleMux = http.NewServeMux()
	exampleServer = httptest.NewServer(exampleMux)
	setupExampleServer()
	ret := m.Run()
	exampleServer.Close()
	os.Exit(ret)
}
