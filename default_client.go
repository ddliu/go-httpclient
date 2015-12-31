// Copyright 2014-2015 Liu Dong <ddliuhb@gmail.com>.
// Licensed under the MIT license.

// Powerful and easy to use http client
package httpclient

import (
	"io"
	"net/http"
)

// The default client for convenience
var defaultClient *HttpClient

var Defaults func(Map) *HttpClient
var Begin func() *HttpClient
var Do func(string, string, map[string]string, io.Reader) (*Response, error)
var Get func(string, map[string]string) (*Response, error)
var Delete func(string, map[string]string) (*Response, error)
var Head func(string, map[string]string) (*Response, error)
var Post func(string, map[string]string) (*Response, error)
var PostMultipart func(string, map[string]string) (*Response, error)
var WithOption func(int, interface{}) *HttpClient
var WithOptions func(Map) *HttpClient
var WithHeader func(string, string) *HttpClient
var WithHeaders func(map[string]string) *HttpClient
var WithCookie func(...*http.Cookie) *HttpClient
var Cookies func(string) []*http.Cookie
var CookieValues func(string) map[string]string
var CookieValue func(string, string) string

func init() {
	defaultClient = NewHttpClient()
	Defaults = defaultClient.Defaults
	Begin = defaultClient.Begin
	Do = defaultClient.Do
	Get = defaultClient.Get
	Delete = defaultClient.Delete
	Head = defaultClient.Head
	Post = defaultClient.Post
	PostMultipart = defaultClient.PostMultipart
	WithOption = defaultClient.WithOption
	WithOptions = defaultClient.WithOptions
	WithHeader = defaultClient.WithHeader
	WithHeaders = defaultClient.WithHeaders
	WithCookie = defaultClient.WithCookie
	Cookies = defaultClient.Cookies
	CookieValues = defaultClient.CookieValues
	CookieValue = defaultClient.CookieValue
}
