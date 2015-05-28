// Copyright (c) 2015, Álvaro Vilanova Vidal
// Use of this source code is governed by a BSD 3-Clause license that can be
// found in the LICENSE file.

package dropbox

import "net/http"

type Entry struct {
	Name string `json:"name"`
}

func (s *FilesService) ListFolder(path string) (entries []Entry, resp *http.Response, err error) {
	cursor := ""
	for {
		var ents []Entry
		ents, cursor, resp, err = s.listFolder(path, cursor)
		if err != nil {
			return
		}
		entries = append(entries, ents...)
		if cursor == "" {
			return
		}
	}
}

func (s *FilesService) listFolder(path, cursor string) (entries []Entry, nextCursor string, resp *http.Response, err error) {
	params := struct {
		Path   string `json:"path"`
		Cursor string `json:"cursor,omitempty"`
	}{
		Path:   path,
		Cursor: cursor,
	}
	req, err := s.client.NewRPCRequest("POST", "2-beta/files/list_folder", &params)
	if err != nil {
		return
	}

	var respData struct {
		Entries []Entry `json:"entries"`
		Footer  struct {
			Cursor  string `json:"cursor,omitempty"`
			HasMore bool   `json:"has_more"`
		} `json:"footer"`
	}

	resp, err = s.client.DoRPC(req, &respData)
	if err != nil {
		return
	}
	if respData.Footer.HasMore {
		nextCursor = respData.Footer.Cursor
	}
	entries = respData.Entries
	return
}
