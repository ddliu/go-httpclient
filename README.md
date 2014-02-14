# go-httpclient

[![Build Status](https://travis-ci.org/ddliu/go-httpclient.png)](https://travis-ci.org/ddliu/go-httpclient)

Advanced HTTP client for golang.

## Features

- Simple API
- Direct file upload
- Timeout
- Proxy(HTTP, <del>SOCKS5, SOCKS4, SOCKS4A</del>)
- Redirect Policy

## Why

There is already a built in http module for golang, and it's possible to achieve any tasks with it. But I found it painful to do some simple tasks. For example, set timeout, set proxy, upload files etc.

I used to be using CURL, and have built a PHP library on top of it. It turned out solving my problems. So I decided to built one for golang.

## Installation

```bash
go get github.com/ddliu/go-httpclient
```

## Usage

In most cases you just need the `Get` and `Post` method

- `func (this *HttpClient) Get(url string, params map[string]string, headers map[string]string, options map[int]interface{}) (*http.Response, error)`
- `func (this *HttpClient) Post(url string, params map[string]string, headers map[string]string, isMultipart bool, options map[int]interface{}) (*http.Response, error)`

```go
package main

import (
    "github.com/ddliu/go-httpclient"
    "fmt"
)

func main() {
    c := httpclient.NewHttpClient(nil)
    res, err := c.Get("http://google.com/search", map[string]string{
        "q": "news",
    }, nil, nil)

    fmt.Println(res.StatusCode, err)

    res, err := c.Post("http://dropbox.com/upload", map[string]string {
        "@file": "/tmp/hello.pdf",
    }, nil, true, map[int]interface{} {
        httpclient.TIMEOUT: 60,
    })

    fmt.Println(res, err)
}
```

## Options

You can specify default options for the `httpclient` instance, and then override it by passing new options for each request.

Available options as below:

- `OPT_FOLLOWLOCATION`: TRUE to follow any "Location: " header that the server sends as part of the HTTP header
- `OPT_CONNECTTIMEOUT`: The number of seconds to wait while trying to connect. Use 0 to wait indefinitely.
- `OPT_CONNECTTIMEOUT_MS`: The number of milliseconds to wait while trying to connect. Use 0 to wait indefinitely.
- `OPT_MAXREDIRS`: The maximum amount of HTTP redirections to follow. Use this option alongside `OPT_FOLLOWLOCATION`.
- `OPT_PROXYTYPE`: Specify the proxy type. Valid options are `PROXY_HTTP`, `PROXY_SOCKS4`, `PROXY_SOCKS5`, `PROXY_SOCKS4A`. Only `PROXY_HTTP` is supported currently. 
- `OPT_TIMEOUT`: The maximum number of seconds to allow httpclient functions to execute.
- `OPT_TIMEOUT_MS`: The maximum number of milliseconds to allow httpclient functions to execute.
- `OPT_COOKIEJAR`: TODO
- `OPT_INTERFACE`: TODO
- `OPT_PROXY`: Proxy host and port(127.0.0.1:1080).
- `OPT_REFERER`: The `Referer` header of the request.
- `OPT_USERAGENT`: The `User-Agent` header of the request.
- `OPT_REDIRECT_POLICY`: Function to check redirect.
- `OPT_PROXY_FUNC`: Function to specify proxy.

## TODO

- Socks proxy support
- COOKIE
- API improvement

## Changelog

### v0.1.0 (2014-02-14)

Initial release