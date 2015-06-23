// Copyright (c) 2015, Álvaro Vilanova Vidal
// Use of this source code is governed by a BSD 3-Clause license that can be
// found in the LICENSE file.

package dropbox

import "net/http"

// AccountInfo contains a Dropbox user account information.
type AccountInfo struct {
	// The user's unique Dropbox ID.
	ID string `json:"account_id"`

	// The user's name
	Name Username `json:"name"`

	// The user's e-mail.
	Email string `json:"email"`

	// The user's two-letter country code, if available.
	Country string `json:"country"`

	// Locale preference set by the user (e.g. en-us).
	Locale string `json:"locale"`

	// The user's referral link.
	ReferralLink string `json:"referral_link"`

	// If true, there is a paired account associated with this user.
	IsPaired bool `json:"is_paired"`
}

// Username contains information about an user name.
type Username struct {
	// The user's given name.
	GivenName string `json:"given_name"`

	// The user's surname.
	Surname string `json:"surname"`

	// The locale-dependent familiar name for the user.
	FamiliarName string `json:"familiar_name"`

	// The user's display name.
	DisplayName string `json:"display_name"`
}

// GetAccount retrieves information about the current user account.
func (s *UsersService) GetAccount() (*AccountInfo, *http.Response, error) {
	req, err := s.client.NewRPCRequest("POST", "2-beta/users/get_current_account", nil)
	if err != nil {
		return nil, nil, err
	}

	var info AccountInfo
	resp, err := s.client.DoRPC(req, &info)
	if err != nil {
		return nil, resp, err
	}

	return &info, resp, nil
}
