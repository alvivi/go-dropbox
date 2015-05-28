// Copyright (c) 2015, Álvaro Vilanova Vidal
// Use of this source code is governed by a BSD 3-Clause license that can be
// found in the LICENSE file.

package dropbox

// FilesService handles communication with the files and metadata related
// methods of the Dropbox API.
type FilesService struct {
	client *Client
}
