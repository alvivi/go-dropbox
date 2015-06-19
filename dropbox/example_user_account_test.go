// Copyright (c) 2015, Álvaro Vilanova Vidal
// Use of this source code is governed by a BSD 3-Clause license that can be
// found in the LICENSE file.

package dropbox

import "fmt"

func ExampleUsersService_GetAccount() {
	// Use golang.org/x/oauth2 for authentication:
	// ts := oauth2.StaticTokenSource(&oauth2.Token{AccessToken: ACCESS_TOKEN})
	// tc := oauth2.NewClient(oauth2.NoContext, ts)
	// c := dropbox.NewClient(tc)
	c := dropbox.NewClient(nil)

	account, _, err := c.Users.GetAccount()
	if err != nil {
		fmt.Printf("Error: %s\n", err.Error())
		return
	}
	fmt.Printf("Hello %s!\n", account.Name.DisplayName)

	// Output:
	// Hello Drew!
}
