// Copyright 2014-2019 Liu Dong <ddliuhb@gmail.com>.
// Licensed under the MIT license.

package main

import (
	"fmt"
	"github.com/ddliu/go-httpclient"
	"net/http"
)

const (
	USERAGENT       = "my awsome httpclient"
	TIMEOUT         = 30
	CONNECT_TIMEOUT = 5
	SERVER          = "https://github.com"
)

func main() {
	httpclient.Defaults(httpclient.Map{
		"opt_useragent":   USERAGENT,
		"opt_timeout":     TIMEOUT,
		"Accept-Encoding": "gzip,deflate,sdch",
	})

	res, _ := httpclient.
		WithHeader("Accept-Language", "en-us").
		WithCookie(&http.Cookie{
			Name:  "name",
			Value: "github",
		}).
		WithHeader("Referer", "http://google.com").
		Get(SERVER)

	fmt.Println("Cookies:")
	for k, v := range httpclient.CookieValues(SERVER) {
		fmt.Println(k, ":", v)
	}

	fmt.Println("Response:")
	fmt.Println(res.ToString())
}
