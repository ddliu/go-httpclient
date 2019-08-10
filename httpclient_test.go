// Copyright 2014-2019 Liu Dong <ddliuhb@gmail.com>.
// Licensed under the MIT license.

// Test httpclient with httpbin(http://httpbin.org)
package httpclient

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"strings"
	"testing"
)

// common response format on httpbin.org
type ResponseInfo struct {
	Gzipped   bool              `json:"gzipped"`
	Method    string            `json:"method"`
	Origin    string            `json:"origin"`
	Useragent string            `json:"user-agent"` // http://httpbin.org/user-agent
	Form      map[string]string `json:"form"`
	Files     map[string]string `json:"files"`
	Headers   map[string]string `json:"headers"`
	Cookies   map[string]string `json:"cookies"`
}

func TestRequest(t *testing.T) {
	// get
	res, err := NewHttpClient().
		Get("http://httpbin.org/get")

	if err != nil {
		t.Error("get failed", err)
	}

	if res.StatusCode != 200 {
		t.Error("Status Code not 200")
	}

	// post
	res, err = NewHttpClient().
		Post("http://httpbin.org/post", map[string]string{
			"username": "dong",
			"password": "******",
		})

	if err != nil {
		t.Error("post failed", err)
	}

	if res.StatusCode != 200 {
		t.Error("Status Code not 200")
	}

	body, err := res.ReadAll()

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
	res, err = NewHttpClient().
		Post("http://httpbin.org/post", map[string]string{
			"message": "Hello world!",
			"@image":  "README.md",
		})

	if err != nil {
		t.Error(err)
	}

	if res.StatusCode != 200 {
		t.Error("Status Code is not 200")
	}

	body, err = res.ReadAll()

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

func TestResponse(t *testing.T) {
	c := NewHttpClient()
	res, err := c.
		Get("http://httpbin.org/user-agent")

	if err != nil {
		t.Error(err)
	}

	// read with ioutil
	defer res.Body.Close()
	body1, err := ioutil.ReadAll(res.Body)

	if err != nil {
		t.Error(err)
	}

	res, err = c.
		Get("http://httpbin.org/user-agent")

	if err != nil {
		t.Error(err)
	}

	body2, err := res.ReadAll()

	res, err = c.
		Get("http://httpbin.org/user-agent")

	if err != nil {
		t.Error(err)
	}

	body3, err := res.ToString()

	if err != nil {
		t.Error(err)
	}
	if string(body1) != string(body2) || string(body1) != body3 {
		t.Error("Error response body")
	}
}

func TestHead(t *testing.T) {
	c := NewHttpClient()
	res, err := c.Head("http://httpbin.org/get")
	if err != nil {
		t.Error(err)
	}

	if res.StatusCode != 200 {
		t.Error("Status code is not 200")
	}

	if body, err := res.ToString(); err != nil || body != "" {
		t.Error("HEAD should not get body")
	}

}

func TestDelete(t *testing.T) {
	c := NewHttpClient()
	res, err := c.Delete("http://httpbin.org/delete")
	if err != nil {
		t.Error(err)
	}

	if res.StatusCode != 200 {
		t.Error("Status code is not 200")
	}
}

func TestOptions(t *testing.T) {
	c := NewHttpClient()
	res, err := c.Options("http://httpbin.org")
	if err != nil {
		t.Error(err)
	}

	if res.StatusCode != 200 {
		t.Errorf("Status code is not 200: %d", res.StatusCode)
	}
}

func TestPatch(t *testing.T) {
	c := NewHttpClient()
	res, err := c.Patch("http://httpbin.org/patch")
	if err != nil {
		t.Error(err)
	}

	if res.StatusCode != 200 {
		t.Errorf("Status code is not 200: %d", res.StatusCode)
	}
}

func TestPostJson(t *testing.T) {
	c := NewHttpClient()
	type jsonDataType struct {
		Name string
	}

	jsonData := jsonDataType{
		Name: "httpclient",
	}

	res, err := c.PostJson("http://httpbin.org/post", jsonData)
	if err != nil {
		t.Error(err)
	}

	if res.StatusCode != 200 {
		t.Error("Status code is not 200")
	}
}

func TestPutJson(t *testing.T) {
	c := NewHttpClient()
	type jsonDataType struct {
		Name string
	}

	jsonData := jsonDataType{
		Name: "httpclient",
	}

	res, err := c.PutJson("http://httpbin.org/put", jsonData)
	if err != nil {
		t.Error(err)
	}

	if res.StatusCode != 200 {
		t.Error("Status code is not 200")
	}
}

func TestPatchJson(t *testing.T) {
	c := NewHttpClient()
	type jsonDataType struct {
		Name string
	}

	jsonData := jsonDataType{
		Name: "httpclient",
	}

	res, err := c.PatchJson("http://httpbin.org/patch", jsonData)
	if err != nil {
		t.Error(err)
	}

	if res.StatusCode != 200 {
		t.Error("Status code is not 200")
	}
}

func TestHeaders(t *testing.T) {
	// set referer in options
	res, err := NewHttpClient().
		WithHeader("header1", "value1").
		WithOption(OPT_REFERER, "http://google.com").
		Get("http://httpbin.org/get")

	if err != nil {
		t.Error(err)
	}

	var info ResponseInfo

	body, err := res.ReadAll()

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
	if !ok || useragent != USERAGENT {
		t.Error("useragent is not set properly")
	}

	value, ok := info.Headers["Header1"]
	if !ok || value != "value1" {
		t.Error("custom header is not set properly")
	}
}

func _TestProxy(t *testing.T) {
	proxy := "127.0.0.1:1080"

	res, err := NewHttpClient().
		WithOption(OPT_PROXY, proxy).
		Get("http://httpbin.org/get")

	if err != nil {
		t.Error(err)
	}

	if res.StatusCode != 200 {
		t.Error("StatusCode is not 200")
	}

	res, err = NewHttpClient().
		WithOption(OPT_PROXY_FUNC, func(*http.Request) (int, string, error) {
			return PROXY_HTTP, proxy, nil
		}).
		Get("http://httpbin.org/get")

	if err != nil {
		t.Error(err)
	}

	if res.StatusCode != 200 {
		t.Error("StatusCode is not 200")
	}
}

func TestTimeout(t *testing.T) {
	// connect timeout
	res, err := NewHttpClient().
		WithOption(OPT_CONNECTTIMEOUT_MS, 1).
		Get("http://httpbin.org/get")

	if err == nil {
		t.Error("OPT_CONNECTTIMEOUT_MS does not work")
	}

	if !IsTimeoutError(err) {
		t.Error("Maybe it's not a timeout error?", err)
	}

	// timeout
	res, err = NewHttpClient().
		WithOption(OPT_TIMEOUT, 3).
		Get("http://httpbin.org/delay/3")

	if err == nil {
		t.Error("OPT_TIMEOUT does not work")
	}

	if !strings.Contains(err.Error(), "timeout") {
		t.Error("Maybe it's not a timeout error?", err)
	}

	// no timeout
	res, err = NewHttpClient().
		WithOption(OPT_TIMEOUT, 100).
		Get("http://httpbin.org/delay/3")

	if err != nil {
		t.Error("OPT_TIMEOUT does not work properly")
	}

	if res.StatusCode != 200 {
		t.Error("StatusCode is not 200")
	}
}

func TestRedirect(t *testing.T) {
	c := NewHttpClient().Defaults(Map{
		OPT_USERAGENT: "test redirect",
	})
	// follow locatioin
	res, err := c.
		WithOptions(Map{
			OPT_FOLLOWLOCATION: true,
			OPT_MAXREDIRS:      10,
		}).
		Get("http://httpbin.org/redirect/3")

	if err != nil {
		t.Error(err)
	}

	if res.StatusCode != 200 || res.Request.URL.String() != "http://httpbin.org/get" {
		t.Error("Redirect failed")
	}

	// should keep useragent
	var info ResponseInfo

	body, err := res.ReadAll()

	if err != nil {
		t.Error(err)
	}

	err = json.Unmarshal(body, &info)

	if err != nil {
		t.Error(err)
	}

	if useragent, ok := info.Headers["User-Agent"]; !ok || useragent != "test redirect" {
		t.Error("Useragent is not passed through")
	}

	// no follow
	res, err = c.
		WithOption(OPT_FOLLOWLOCATION, false).
		Get("http://httpbin.org/relative-redirect/3")

	if err == nil {
		t.Error("Must not follow location")
	}

	if !strings.Contains(err.Error(), "redirect not allowed") {
		t.Error(err)
	}

	if res.StatusCode != 302 || res.Header.Get("Location") != "/relative-redirect/2" {
		t.Error("Redirect failed: ", res.StatusCode, res.Header.Get("Location"))
	}

	// maxredirs
	res, err = c.
		WithOption(OPT_MAXREDIRS, 2).
		Get("http://httpbin.org/relative-redirect/3")

	if err == nil {
		t.Error("Must not follow through")
	}

	if !IsRedirectError(err) {
		t.Error("Not a redirect error", err)
	}

	if !strings.Contains(err.Error(), "stopped after 2 redirects") {
		t.Error(err)
	}

	if res.StatusCode != 302 || res.Header.Get("Location") != "/relative-redirect/1" {
		t.Error("OPT_MAXREDIRS does not work properly")
	}

	// custom redirect policy
	res, err = c.
		WithOption(OPT_REDIRECT_POLICY, func(req *http.Request, via []*http.Request) error {
			if req.URL.String() == "http://httpbin.org/relative-redirect/1" {
				return fmt.Errorf("should stop here")
			}

			return nil
		}).
		Get("http://httpbin.org/relative-redirect/3")

	if err == nil {
		t.Error("Must not follow through")
	}

	if !strings.Contains(err.Error(), "should stop here") {
		t.Error(err)
	}

	if res.StatusCode != 302 || res.Header.Get("Location") != "/relative-redirect/1" {
		t.Error("OPT_REDIRECT_POLICY does not work properly")
	}
}

func TestCookie(t *testing.T) {
	c := NewHttpClient()

	res, err := c.
		WithCookie(&http.Cookie{
			Name:  "username",
			Value: "dong",
		}).
		Get("http://httpbin.org/cookies")

	if err != nil {
		t.Error(err)
	}

	body, err := res.ReadAll()

	if err != nil {
		t.Error(err)
	}

	var info ResponseInfo

	err = json.Unmarshal(body, &info)

	if err != nil {
		t.Error(err)
	}

	if username, ok := info.Cookies["username"]; !ok || username != "dong" {
		t.Error("cookie is not set properly")
	}

	if c.CookieValue("http://httpbin.org/cookies", "username") != "dong" {
		t.Error("cookie is not set properly")
	}

	// get old cookie
	res, err = c.
		Get("http://httpbin.org/cookies", nil)

	if err != nil {
		t.Error(err)
	}

	body, err = res.ReadAll()

	if err != nil {
		t.Error(err)
	}

	err = json.Unmarshal(body, &info)

	if err != nil {
		t.Error(err)
	}

	if username, ok := info.Cookies["username"]; !ok || username != "dong" {
		t.Error("cookie lost")
	}

	if c.CookieValue("http://httpbin.org/cookies", "username") != "dong" {
		t.Error("cookie lost")
	}

	// update cookie
	res, err = c.
		WithCookie(&http.Cookie{
			Name:  "username",
			Value: "octcat",
		}).
		Get("http://httpbin.org/cookies")

	if err != nil {
		t.Error(err)
	}

	body, err = res.ReadAll()

	if err != nil {
		t.Error(err)
	}

	err = json.Unmarshal(body, &info)

	if err != nil {
		t.Error(err)
	}

	if username, ok := info.Cookies["username"]; !ok || username != "octcat" {
		t.Error("cookie update failed")
	}

	if c.CookieValue("http://httpbin.org/cookies", "username") != "octcat" {
		t.Error("cookie update failed")
	}
}

func TestGzip(t *testing.T) {
	c := NewHttpClient()
	res, err := c.
		WithHeader("Accept-Encoding", "gzip, deflate").
		Get("http://httpbin.org/gzip")

	if err != nil {
		t.Error(err)
	}

	body, err := res.ReadAll()

	if err != nil {
		t.Error(err)
	}

	var info ResponseInfo

	err = json.Unmarshal(body, &info)

	if err != nil {
		t.Error(err)
	}

	if !info.Gzipped {
		t.Error("Parse gzip failed")
	}
}

func _TestCurrentUA(ch chan bool, t *testing.T, c *HttpClient, ua string) {
	res, err := c.
		Begin().
		WithOption(OPT_USERAGENT, ua).
		Get("http://httpbin.org/headers")

	if err != nil {
		t.Error(err)
	}

	body, err := res.ReadAll()

	if err != nil {
		t.Error(err)
	}

	var info ResponseInfo
	err = json.Unmarshal(body, &info)

	if err != nil {
		t.Error(err)
	}

	if resUA, ok := info.Headers["User-Agent"]; !ok || resUA != ua {
		t.Error("TestCurrentUA failed")
	}

	ch <- true
}

func TestConcurrent(t *testing.T) {
	total := 100
	chs := make([]chan bool, total)
	c := NewHttpClient()
	for i := 0; i < total; i++ {
		chs[i] = make(chan bool)
		go _TestCurrentUA(chs[i], t, c, fmt.Sprint("go-httpclient UA-", i))
	}

	for _, ch := range chs {
		<-ch
	}
}

func TestIssue10(t *testing.T) {
	var testString = "gpThzrynEC1MdenWgAILwvL2CYuNGO9RwtbH1NZJ1GE31ywFOCY%2BLCctUl86jBi8TccpdPI5ppZ%2Bgss%2BNjqGHg=="
	c := NewHttpClient()
	res, err := c.Post("http://httpbin.org/post", map[string]string{
		"a": "a",
		"b": "b",
		"c": testString,
		"d": "d",
	})

	if err != nil {
		t.Error(err)
	}

	body, err := res.ReadAll()

	if err != nil {
		t.Error(err)
	}

	var info ResponseInfo

	err = json.Unmarshal(body, &info)

	if err != nil {
		t.Error(err)
	}

	if info.Form["c"] != testString {
		t.Error("error")
	}
}

func TestOptDebug(t *testing.T) {
	c := NewHttpClient()
	c.
		WithOption(OPT_DEBUG, true).
		Get("http://httpbin.org/get")
}

func TestUnsafeTLS(t *testing.T) {
	unsafeUrl := "https://expired.badssl.com/"
	c := NewHttpClient()
	_, err := c.
		Get(unsafeUrl, nil)
	if err == nil {
		t.Error("Unexcepted unsafe url:" + unsafeUrl)
	}

	res, err := c.
		WithOption(OPT_UNSAFE_TLS, true).
		Get(unsafeUrl)
	if err != nil {
		t.Error(err)
	}

	if res.StatusCode != 200 {
		t.Error("OPT_UNSAFE_TLS error")
	}
}

func TestPutJsonWithCharset(t *testing.T) {
	c := NewHttpClient()
	type jsonDataType struct {
		Name string
	}

	jsonData := jsonDataType{
		Name: "httpclient",
	}

	contentType := "application/json; charset=utf-8"
	res, err := c.
		WithHeader("Content-Type", contentType).
		PutJson("http://httpbin.org/put", jsonData)
	if err != nil {
		t.Error(err)
	}

	body, err := res.ReadAll()

	if err != nil {
		t.Error(err)
	}

	var info ResponseInfo

	err = json.Unmarshal(body, &info)

	if err != nil {
		t.Error(err)
	}

	if info.Headers["Content-Type"] != contentType {
		t.Error("Setting charset not working: " + info.Headers["Content-Type"])
	}
}
