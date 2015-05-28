// Copyright (c) 2015, Álvaro Vilanova Vidal
// Use of this source code is governed by a BSD 3-Clause license that can be
// found in the LICENSE file.

// This example shows how to list a Dropbox folder.
package main

import (
	"flag"
	"fmt"

	"github.com/alvivi/go-dropbox/dropbox"
	"golang.org/x/oauth2"
)

var token = flag.String("token", "", "Dropbox OAuth2 bearer token")
var path = flag.String("path", "/", "Folder to list")

func init() {
	flag.Parse()
}

func main() {
	if token == nil || *token == "" {
		flag.PrintDefaults()
		return
	}
	ts := oauth2.StaticTokenSource(&oauth2.Token{AccessToken: *token})
	tc := oauth2.NewClient(oauth2.NoContext, ts)
	c := dropbox.NewClient(tc)

	if *path == "/" {
		*path = ""
	}

	entries, _, err := c.Files.ListFolder(*path)
	if err != nil {
		fmt.Println(err.Error())
		return
	}

	for _, entry := range entries {
		fmt.Println(entry.Name)
	}
}
