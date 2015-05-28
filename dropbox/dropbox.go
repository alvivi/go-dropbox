// Copyright (c) 2015, Álvaro Vilanova Vidal
// Use of this source code is governed by a BSD 3-Clause license that can be
// found in the LICENSE file.

package dropbox

import (
	"bytes"
	"encoding/json"
	"io"
	"io/ioutil"
	"net/http"
	"net/url"
	"strings"
)

const (
	libraryVersion    = "0.1"
	defaultBaseURL    = "https://api.dropbox.com/"
	defaultContentURL = "https://api-content.dropbox.com/"
	userAgent         = "go-dropbox/" + libraryVersion

	defaultMediaType = "application/json; charset=utf-8"
)

// RPCRequest is a RPC Style Request. The request and response bodies are both
// JSON.
type RPCRequest http.Request

// A Client manages communication with the Dropbox API.
type Client struct {
	// HTTP client used to communicate with the API.
	client *http.Client

	// Base URL for API requests. Defaults to the public Dropbox API.
	BaseURL *url.URL

	// Base URL for content API request. Defaults to the public Dropbox content
	// API.
	ContentURL *url.URL

	// User agent used when communicating with the Dropbox API.
	UserAgent string

	// Used to specify language settings for user error messages and other
	// language specific text. If your app supports any language other than
	// English, insert the appropriate IETF language tag.
	//
	// More info at https://www.dropbox.com/developers/core/docs#param.locale
	Locale string

	// Services used for talking to different parts of the Dropbox API.
	Users *UsersService
	Files *FilesService
}

// NewClient returns a new Dropbox API client. If a nil httpClient is provided,
// http.DefaultClient will be used. To use API methods which require
// authentication, provide an http.Client that will perform the authentication
// for you (such as that provided by the golang.org/x/oauth2 library).
func NewClient(httpClient *http.Client) *Client {
	if httpClient == nil {
		httpClient = http.DefaultClient
	}
	baseURL, _ := url.Parse(defaultBaseURL)
	contentURL, _ := url.Parse(defaultContentURL)
	c := &Client{
		client:     httpClient,
		BaseURL:    baseURL,
		ContentURL: contentURL,
		UserAgent:  userAgent,
	}

	c.Users = &UsersService{c}
	c.Files = &FilesService{c}

	return c
}

// NewRPCRequest returns a new RPC style request. A relative URL can be provided
// in urlStr, in which case it is resolved relative to the BaseURL of the
// Client. Relative URLs should always be specified without a preceding slash.
// Body, if specified, must be a valid JSON marshable value.
func (c *Client) NewRPCRequest(method, urlStr string, body interface{}) (*RPCRequest, error) {
	req, err := c.newRequest(method, urlStr, func(w io.Writer) error {
		return json.NewEncoder(w).Encode(body)
	})
	if err != nil {
		return nil, err
	}
	req.Header.Add("Accept", "application/json; charset=utf-8")
	req.Header.Add("Content-Type", "application/json; charset=utf-8")
	return (*RPCRequest)(req), nil
}

// DoRPC sends a RPC style request and returns the API response. The API
// response is JSON decoded and stored in the value pointed to by v, or returned
// as an error if an API error has occurred.
func (c *Client) DoRPC(req *RPCRequest, v interface{}) (*http.Response, error) {
	resp, err := c.client.Do((*http.Request)(req))
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	err = checkResponse(resp)
	if err != nil {
		return resp, err
	}

	if v == nil {
		return resp, err
	}

	err = json.NewDecoder(resp.Body).Decode(v)
	return resp, err
}

// UnexpectedError is an error returned by go-dropbox when no more information
// is provided.
type UnexpectedError struct{}

func (e UnexpectedError) Error() string {
	return "unexpected error"
}

// Error is a Dropbox API error.
type Error struct {
	Reason string `json:"reason"`
}

func (e Error) Error() string {
	return e.Reason
}

func checkContentType(res *http.Response, ctype string) bool {
	ctypeHeaderStr := strings.TrimSpace(res.Header.Get("Content-Type"))
	return strings.HasPrefix(ctypeHeaderStr, ctype)
}

// note that this method do not close response body
func checkResponse(res *http.Response) error {
	if c := res.StatusCode; 200 <= c && c <= 299 {
		return nil
	}
	if checkContentType(res, "application/json") {
		var dpErr Error
		err := json.NewDecoder(res.Body).Decode(&dpErr)
		if err != nil {
			return err
		}
		return dpErr
	}
	if checkContentType(res, "text/plain") {
		buf, err := ioutil.ReadAll(res.Body)
		if err != nil {
			return err
		}
		return Error{string(buf)}
	}
	return UnexpectedError{}
}

func (c *Client) newRequest(method, urlStr string, bw func(io.Writer) error) (*http.Request, error) {
	rel, err := url.Parse(urlStr)
	if err != nil {
		return nil, err
	}
	u := c.BaseURL.ResolveReference(rel)

	var buffer io.ReadWriter
	if bw != nil {
		buffer = new(bytes.Buffer)
		if err := bw(buffer); err != nil {
			return nil, err
		}
	}

	req, err := http.NewRequest(method, u.String(), buffer)
	if err != nil {
		return nil, err
	}

	if c.UserAgent != "" {
		req.Header.Add("User-Agent", c.UserAgent)
	}

	return req, nil
}
