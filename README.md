# go-httpclient

[![Build Status](https://travis-ci.org/ddliu/go-httpclient.png)](https://travis-ci.org/ddliu/go-httpclient)

Advanced HTTP client for golang.

## Features

- Simple API
- Direct file upload
- Timeout
- Proxy(HTTP, <del>SOCKS5, SOCKS4, SOCKS4A</del>)
- Cookie
- Redirect Policy

## Why

There is already a built in http module for golang, and it's possible to achieve any tasks with it. But I found it painful to do some simple tasks. For example, set timeout, set proxy, upload files etc.

I used to be using CURL, and have built a PHP library on top of it. It turned out solving my problems. So I decided to built one for golang.

## Installation

```bash
go get github.com/ddliu/go-httpclient
```

## Usage

### Initialize

Use `NewHttpClient` to create a client.

```
package main

import (
    "github.com/ddliu/go-httpclient"
)

var c := httpclient.NewHttpClient(map[int]interface{} {
    httpclient.OPT_USERAGENT: "my awsome httpclient",
})
```

The `httpclient.OPT_XXX` options define basic behaviours of this client, they are shared between different HTTP requests.

### GET/POST

In most cases you just need the `Get` and `Post` method after initializing:

```
func (this *HttpClient) Get(url string, params map[string]string) (*httpclient.Response, error)

func (this *HttpClient) Post(url string, params map[string]string) (*httpclient.Response, error)
```

```go
    res, err := c.Get("http://google.com/search", map[string]string{
        "q": "news",
    })

    fmt.Println(res.StatusCode, err)

    // post file
    res, err := c.Post("http://dropbox.com/upload", map[string]string {
        "@file": "/tmp/hello.pdf",
    })

    fmt.Println(res, err)
```

### Customize Request

Before you start a new HTTP request with `Get` or `Post` method, you can specify options, headers or cookies.

```go
c.
    WithHeader("User-Agent", "Super Robot").
    WithHeader("custom-header", "value").
    WithHeaders(map[string]string {
        "another-header": "another-value",
        "and-another-header": "another-value",
    }).
    WithOption(httpclent.OPT_TIMEOUT, 60).
    WithCookie(&http.Cookie{
        Name: "uid",
        Value: "123",
    }).
    Get("http://github.com", nil)
```

### Response

The `httpclient.Response` is a thin wrap of `http.Response`.

```go
// traditional
res, err := c.Get("http://google.com", nil)
bodyBytes, err := ioutil.ReadAll(res.Body)
res.Body.Close()

// ToString
res, err = c.Get("http://google.com", nil)
bodyString := res.ToString()

// ReadAll
res, err = c.Get("http://google.com", nil)
bodyBytes := res.ReadAll()
```

### Handle Cookies

```
url := "http://github.com"
c.
    WithCookie(&http.Cookie{
        Name: "uid",
        Value: "123",
    }).
    Get(url, nil)

for _, cookie := range c.Cookies() {
    fmt.Println(c.Name, c.Value)
}

for k, v := range c.CookieValues() {
    fmt.Println(k, v)
}

fmt.Println(c.CookieValue("uid"))
```

### Concurrent Safe

If you've created one client and want to start many requests concurrently, remember to call the `Begin` method when you begin:

```go
c := httpclient.NewHttpClient(nil)

go func() {
    c.
        Begin().
        WithHeader("Req-A", "a").
        Get("http://google.com")
}()
go func() {
    c.
        Begin().
        WithHeader("Req-B", "b").
        Get("http://google.com")
}()

```

### Full Example

See `examples/main.go`

## Options

Available options as below:

- `OPT_FOLLOWLOCATION`: TRUE to follow any "Location: " header that the server sends as part of the HTTP header. Default to `true`.
- `OPT_CONNECTTIMEOUT`: The number of seconds to wait while trying to connect. Use 0 to wait indefinitely.
- `OPT_CONNECTTIMEOUT_MS`: The number of milliseconds to wait while trying to connect. Use 0 to wait indefinitely.
- `OPT_MAXREDIRS`: The maximum amount of HTTP redirections to follow. Use this option alongside `OPT_FOLLOWLOCATION`.
- `OPT_PROXYTYPE`: Specify the proxy type. Valid options are `PROXY_HTTP`, `PROXY_SOCKS4`, `PROXY_SOCKS5`, `PROXY_SOCKS4A`. Only `PROXY_HTTP` is supported currently. 
- `OPT_TIMEOUT`: The maximum number of seconds to allow httpclient functions to execute.
- `OPT_TIMEOUT_MS`: The maximum number of milliseconds to allow httpclient functions to execute.
- `OPT_COOKIEJAR`: Set to `true` to enable the default cookiejar, or you can set to a `http.CookieJar` instance to use a customized jar. Default to `true`.
- `OPT_INTERFACE`: TODO
- `OPT_PROXY`: Proxy host and port(127.0.0.1:1080).
- `OPT_REFERER`: The `Referer` header of the request.
- `OPT_USERAGENT`: The `User-Agent` header of the request. Default to "go-httpclient v{{VERSION}}".
- `OPT_REDIRECT_POLICY`: Function to check redirect.
- `OPT_PROXY_FUNC`: Function to specify proxy.

## API

See [godoc](https://godoc.org/github.com/ddliu/go-httpclient).

## TODO

- Socks proxy support
- Return different error types

## Changelog

### v0.1.0 (2014-02-14)

Initial release

### v0.2.0 (2014-02-17)

Rewrite API, make it simple

### v0.2.1 (2014-05-18)

Make `http.Client` reusable

### v0.3.0 (2014-05-20)

API improvements

Cookie support

Concurrent safe

### v0.3.1 (2014-05-20)

Add shortcut for response

### v0.3.2 (2014-05-21)

Fix cookie, add cookie retrieving methods