// Copyright (c) 2015, Álvaro Vilanova Vidal
// Use of this source code is governed by a BSD 3-Clause license that can be
// found in the LICENSE file.

package dropbox

import "fmt"

func ExampleFilesService_ListFolder() {
	// Use golang.org/x/oauth2 for authentication:
	// ts := oauth2.StaticTokenSource(&oauth2.Token{AccessToken: ACCESS_TOKEN})
	// tc := oauth2.NewClient(oauth2.NoContext, ts)
	// c := dropbox.NewClient(tc)
	c := dropbox.NewClient(nil)

	entries, _, err := c.Files.ListFolder("/photos")
	if err != nil {
		fmt.Printf("Error: %s\n", err.Error())
		return
	}
	for _, entry := range entries {
		fmt.Println(entry.Name)
	}

	// Output:
	// James.jpg
	// Mary.jpg
	// Richard.jpg
	// Susan.jpg
}
