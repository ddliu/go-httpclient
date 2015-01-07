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

var Begin func() *HttpClient

var Do func(string, string, map[string]string, io.Reader)(*Response, error)

var Get func(string, map[string]string)(*Response, error)
var Post func(string, map[string]string)(*Response, error)
var PostMultipart func(string, map[string]string)(*Response, error)
var WithOption func(int, interface{}) *HttpClient
var WithOptions func(Map) *HttpClient
var WithHeader func(string, string) *HttpClient
var WithHeaders func(map[string]string) *HttpClient
var WithCookie func(...*http.Cookie) *HttpClient
var Cookies func(string)([]*http.Cookie)
var CookieValues func(string)(map[string]string)
var CookieValue func(string, string)(string)

func init() {
    defaultClient = NewHttpClient(nil)
    Begin = defaultClient.Begin
    Do = defaultClient.Do
    Get = defaultClient.Get
    Post = defaultClient.Post
    PostMultipart = defaultClient.PostMultipart
    Cookies = defaultClient.Cookies
    CookieValues = defaultClient.CookieValues
    CookieValue = defaultClient.CookieValue
}