// Copyright (c) 2015, Álvaro Vilanova Vidal
// Use of this source code is governed by a BSD 3-Clause license that can be
// found in the LICENSE file.

package main

import (
	"flag"
	"fmt"

	"github.com/alvivi/go-dropbox/dropbox"
	"golang.org/x/oauth2"
)

var token = flag.String("token", "", "Dropbox OAuth2 bearer token")

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

	account, _, err := c.Users.GetAccount()
	if err != nil {
		fmt.Printf("Error: %s\n", err.Error())
		return
	}

	fmt.Printf("ID: %s\n"+
		"Name:\n"+
		"\tGiven name: %s\n"+
		"\tSurname: %s\n"+
		"\tFamiliar name: %s\n"+
		"\tDisplay name: %s\n"+
		"Email: %s\n"+
		"Country: %s\n"+
		"Locale: %s\n"+
		"Referral link: %s\n"+
		"Is paired: %t\n", account.ID, account.Name.GivenName, account.Name.Surname,
		account.Name.FamiliarName, account.Name.DisplayName, account.Email,
		account.Country, account.Locale, account.ReferralLink, account.IsPaired)
}
