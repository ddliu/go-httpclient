// Test httpclient with httpbin(http://httpbin.org)

package httpclient

import (
    "testing"
    "fmt"
    "strings"
    "io/ioutil"
    "net/http"
    "encoding/json"
)

// common response format on httpbin.org
type ResponseInfo struct {
    Method string `json:"method"`
    Origin string `json:"origin"`
    Useragent string `json:"user-agent"` // http://httpbin.org/user-agent
    Form map[string]string `json:"form"`
    Files map[string]string `json:"files"`
    Headers map[string]string  `json:"headers"`
}

func TestRequest(t *testing.T) {
    c := NewHttpClient(nil)

    // get
    res, err := c.Get("http://httpbin.org/get", nil, nil, nil)

    if err != nil {
        t.Error("get failed", err)
    }

    if res.StatusCode != 200 {
        t.Error("Status Code not 200")
    }

    // post
    res, err = c.Post("http://httpbin.org/post", map[string]string {
        "username": "dong",
        "password": "******",
    }, nil, false, nil)

    if err != nil {
        t.Error("post failed", err)
    }

    if res.StatusCode != 200 {
        t.Error("Status Code not 200")
    }

    defer res.Body.Close()

    body, err := ioutil.ReadAll(res.Body)

    if err != nil {
        t.Error(err)
    }

    var info ResponseInfo

    err = json.Unmarshal(body, &info)

    if err != nil {
        t.Error(err)
    }

    if username, ok := info.Form["username"]; !ok || username != "dong" {
        t.Error("form data is not set properly")
    }

    // post, multipart
    res, err = c.Post("http://httpbin.org/post", map[string]string {
        "message": "Hello world!",
        "@image": "README.md",
    }, nil, true, nil)

    if err != nil {
        t.Error(err)
    }

    if res.StatusCode != 200 {
        t.Error("Status Code is not 200")
    }

    defer res.Body.Close()

    body, err = ioutil.ReadAll(res.Body)

    if err != nil {
        t.Error(err)
    }

    err = json.Unmarshal(body, &info)

    if err != nil {
        t.Error(err)
    }

    image, ok := info.Files["image"]
    if !ok {
        t.Error("file not uploaded")
    }

    imageContent, err := ioutil.ReadFile("README.md")
    if err != nil {
        t.Error(err)
    }

    if string(imageContent) != image {
        t.Error("file is not uploaded properly")
    }
}

func TestHeaders(t *testing.T) {

    c := NewHttpClient(nil)

    // set referer in options
    res, err := c.Get("http://httpbin.org/get", nil, map[string]string {
        "header1": "value1",
    }, map[int]interface{} {
        OPT_REFERER: "http://google.com",
    })

    if err != nil {
        t.Error(err)
    }

    defer res.Body.Close()

    var info ResponseInfo

    body, err := ioutil.ReadAll(res.Body)

    if err != nil {
        t.Error(err)
    }

    err = json.Unmarshal(body, &info)

    if err != nil {
        t.Error(err)
    }

    referer, ok := info.Headers["Referer"]
    if !ok || referer != "http://google.com" {
        t.Error("referer is not set properly")
    }

    useragent, ok := info.Headers["User-Agent"]
    if !ok  || useragent != USERAGENT {
        t.Error("useragent is not set properly")
    }
    
    value, ok := info.Headers["Header1"]
    if !ok || value != "value1" {
        t.Error("custom header is not set properly")
    }
}

func _TestProxy(t *testing.T) {
    c := NewHttpClient(nil)

    proxy := "127.0.0.1:1080"

    res, err := c.Get("http://httpbin.org/get", nil, nil, map[int]interface{} {
        OPT_PROXY: proxy,
    })

    if err != nil {
        t.Error(err)
    }

    if res.StatusCode != 200 {
        t.Error("StatusCode is not 200")
    }

    res, err = c.Get("http://httpbin.org/get", nil, nil, map[int]interface{} {
        OPT_PROXY_FUNC: func(*http.Request) (int, string, error) {
            return PROXY_HTTP, proxy, nil
        },
    })

    if err != nil {
        t.Error(err)
    }

    if res.StatusCode != 200 {
        t.Error("StatusCode is not 200")
    }
}

func TestTimeout(t *testing.T) {
    c := NewHttpClient(nil)

    // connect timeout
    res, err := c.Get("http://httpbin.org/get", nil, nil, map[int]interface{} {
        OPT_CONNECTTIMEOUT_MS: 1,
    })

    if err == nil {
        t.Error("OPT_CONNECTTIMEOUT_MS does not work")
    }

    if !strings.Contains(err.Error(), "timeout") {
        t.Error("Maybe it's not a timeout error?", err)
    }

    // timeout
    res, err = c.Get("http://httpbin.org/delay/3", nil, nil, map[int]interface{} {
        OPT_TIMEOUT: 3,
    })

    if err == nil {
        t.Error("OPT_TIMEOUT does not work")
    }

    if !strings.Contains(err.Error(), "timeout") {
        t.Error("Maybe it's not a timeout error?", err)
    }

    // no timeout
    res, err = c.Get("http://httpbin.org/delay/3", nil, nil, map[int]interface{} {
        OPT_TIMEOUT: 100,
    })

    if err != nil {
        t.Error("OPT_TIMEOUT does not work properly")
    }

    if res.StatusCode != 200 {
        t.Error("StatusCode is not 200")
    }
}

func TestRedirect(t *testing.T) {
    c := NewHttpClient(nil)

    // follow locatioin
    res, err := c.Get("http://httpbin.org/redirect/3", nil, nil, map[int]interface{} {
        OPT_FOLLOWLOCATION: true,
        OPT_MAXREDIRS: 10,
    })

    if err != nil {
        t.Error(err)
    }

    if res.StatusCode != 200 || res.Request.URL.String() != "http://httpbin.org/get" {
        t.Error("Redirect failed")
    }

    // no follow
    res, err = c.Get("http://httpbin.org/redirect/3", nil, nil, map[int]interface{} {
        OPT_FOLLOWLOCATION: false,
    })

    if err == nil {
        t.Error("Must not follow location")
    }

    if !strings.Contains(err.Error(), "redirect not allowed") {
        t.Error(err)
    }

    if res.StatusCode != 302 || res.Header.Get("Location") != "/redirect/2" {
        t.Error("Redirect failed: ", res.StatusCode, res.Header.Get("Location"))
    }

    // maxredirs
    res, err = c.Get("http://httpbin.org/redirect/3", nil, nil, map[int]interface{} {
        OPT_MAXREDIRS: 2,
    })

    if err == nil {
        t.Error("Must not follow through")
    }

    if !strings.Contains(err.Error(), "stopped after 2 redirects") {
        t.Error(err)
    }

    if res.StatusCode != 302 || res.Header.Get("Location") != "/redirect/1" {
        t.Error("OPT_MAXREDIRS does not work properly")
    }

    // custom redirect policy
    res, err = c.Get("http://httpbin.org/redirect/3", nil, nil, map[int]interface{} {
        OPT_REDIRECT_POLICY: func(req *http.Request, via []*http.Request) error {
            if req.URL.String() == "http://httpbin.org/redirect/1" {
                return fmt.Errorf("should stop here")
            }

            return nil
        },
    })

    if err == nil {
        t.Error("Must not follow through")
    }

    if !strings.Contains(err.Error(), "should stop here") {
        t.Error(err)
    }

    if res.StatusCode != 302 || res.Header.Get("Location") != "/redirect/1" {
        t.Error("OPT_REDIRECT_POLICY does not work properly")
    }
}