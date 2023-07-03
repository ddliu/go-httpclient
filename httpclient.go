// Copyright 2014-2019 Liu Dong <ddliuhb@gmail.com>.
// Licensed under the MIT license.

// Powerful and easy to use http client
package httpclient

import (
	"context"
	"fmt"

	"bytes"
	"strings"

	"time"

	"io"
	"io/ioutil"
	"sync"

	"net"
	"net/http"
	"net/http/cookiejar"
	"net/http/httputil"
	"net/url"

	"crypto/tls"

	"compress/gzip"

	"encoding/json"
	"mime/multipart"
)

// Constants definations
// CURL options, see https://github.com/bagder/curl/blob/169fedbdce93ecf14befb6e0e1ce6a2d480252a3/packages/OS400/curl.inc.in
const (
	VERSION   = "0.7.1"
	USERAGENT = "go-httpclient v" + VERSION
)

const (
	PROXY_HTTP int = iota
	PROXY_SOCKS4
	PROXY_SOCKS5
	PROXY_SOCKS4A

	// CURL like OPT
	OPT_AUTOREFERER
	OPT_FOLLOWLOCATION
	OPT_CONNECTTIMEOUT
	OPT_CONNECTTIMEOUT_MS
	OPT_MAXREDIRS
	OPT_PROXYTYPE
	OPT_TIMEOUT
	OPT_TIMEOUT_MS
	OPT_COOKIEJAR
	OPT_INTERFACE
	OPT_PROXY
	OPT_REFERER
	OPT_USERAGENT

	// Other OPT
	OPT_REDIRECT_POLICY
	OPT_PROXY_FUNC
	OPT_DEBUG
	OPT_UNSAFE_TLS

	OPT_CONTEXT

	OPT_BEFORE_REQUEST_FUNC
)

// String map of options
var CONST = map[string]int{
	"OPT_AUTOREFERER":         OPT_AUTOREFERER,
	"OPT_FOLLOWLOCATION":      OPT_FOLLOWLOCATION,
	"OPT_CONNECTTIMEOUT":      OPT_CONNECTTIMEOUT,
	"OPT_CONNECTTIMEOUT_MS":   OPT_CONNECTTIMEOUT_MS,
	"OPT_MAXREDIRS":           OPT_MAXREDIRS,
	"OPT_PROXYTYPE":           OPT_PROXYTYPE,
	"OPT_TIMEOUT":             OPT_TIMEOUT,
	"OPT_TIMEOUT_MS":          OPT_TIMEOUT_MS,
	"OPT_COOKIEJAR":           OPT_COOKIEJAR,
	"OPT_INTERFACE":           OPT_INTERFACE,
	"OPT_PROXY":               OPT_PROXY,
	"OPT_REFERER":             OPT_REFERER,
	"OPT_USERAGENT":           OPT_USERAGENT,
	"OPT_REDIRECT_POLICY":     OPT_REDIRECT_POLICY,
	"OPT_PROXY_FUNC":          OPT_PROXY_FUNC,
	"OPT_DEBUG":               OPT_DEBUG,
	"OPT_UNSAFE_TLS":          OPT_UNSAFE_TLS,
	"OPT_CONTEXT":             OPT_CONTEXT,
	"OPT_BEFORE_REQUEST_FUNC": OPT_BEFORE_REQUEST_FUNC,
}

// Default options for any clients.
var defaultOptions = map[int]interface{}{
	OPT_FOLLOWLOCATION: true,
	OPT_MAXREDIRS:      10,
	OPT_AUTOREFERER:    true,
	OPT_USERAGENT:      USERAGENT,
	OPT_COOKIEJAR:      true,
	OPT_DEBUG:          false,
}

// These options affect transport, transport may not be reused if you change any
// of these options during a request.
var transportOptions = []int{
	OPT_CONNECTTIMEOUT,
	OPT_CONNECTTIMEOUT_MS,
	OPT_PROXYTYPE,
	OPT_TIMEOUT,
	OPT_TIMEOUT_MS,
	OPT_INTERFACE,
	OPT_PROXY,
	OPT_PROXY_FUNC,
	OPT_UNSAFE_TLS,
}

// These options affect cookie jar, jar may not be reused if you change any of
// these options during a request.
var jarOptions = []int{
	OPT_COOKIEJAR,
}

// Thin wrapper of http.Response(can also be used as http.Response).
type Response struct {
	*http.Response
}

// Read response body into a byte slice.
func (this *Response) ReadAll() ([]byte, error) {
	var reader io.ReadCloser
	var err error
	switch this.Header.Get("Content-Encoding") {
	case "gzip":
		reader, err = gzip.NewReader(this.Body)
		if err != nil {
			return nil, err
		}
	default:
		reader = this.Body
	}

	defer reader.Close()
	return ioutil.ReadAll(reader)
}

// Read response body into string.
func (this *Response) ToString() (string, error) {
	bytes, err := this.ReadAll()
	if err != nil {
		return "", err
	}

	return string(bytes), nil
}

func (this *Response) ToJson(v interface{}) error {
	bytes, err := this.ReadAll()
	if err != nil {
		return err
	}
	if err = json.Unmarshal(bytes, v); err != nil {
		return err
	}
	return nil
}

// Prepare a request.
func prepareRequest(method string, url_ string, headers map[string]string,
	body io.Reader, options map[int]interface{}) (*http.Request, error) {
	req, err := http.NewRequest(method, url_, body)

	if err != nil {
		return nil, err
	}

	// OPT_REFERER
	if referer, ok := options[OPT_REFERER]; ok {
		if refererStr, ok := referer.(string); ok {
			req.Header.Set("Referer", refererStr)
		}
	}

	// OPT_USERAGENT
	if useragent, ok := options[OPT_USERAGENT]; ok {
		if useragentStr, ok := useragent.(string); ok {
			req.Header.Set("User-Agent", useragentStr)
		}
	}

	for k, v := range headers {
		req.Header.Set(k, v)
	}

	return req, nil
}

func prepareTimeout(options map[int]interface{}) (time.Duration, error) {
	var timeout time.Duration

	if timeoutMS_, ok := options[OPT_TIMEOUT_MS]; ok {
		if timeoutMS, ok := timeoutMS_.(int); ok {
			timeout = time.Duration(timeoutMS) * time.Millisecond
		} else {
			return 0, fmt.Errorf("OPT_TIMEOUT_MS must be int")
		}
	} else if timeout_, ok := options[OPT_TIMEOUT]; ok {
		if timeout, ok = timeout_.(time.Duration); !ok {
			if timeoutS, ok := timeout_.(int); ok {
				timeout = time.Duration(timeoutS) * time.Second
			} else {
				return 0, fmt.Errorf("OPT_TIMEOUT must be int or time.Duration")
			}
		}
	}

	return timeout, nil
}

func prepareConnTimeout(options map[int]interface{}) (time.Duration, error) {
	var connectTimeout time.Duration

	if connectTimeoutMS_, ok := options[OPT_CONNECTTIMEOUT_MS]; ok {
		if connectTimeoutMS, ok := connectTimeoutMS_.(int); ok {
			connectTimeout = time.Duration(connectTimeoutMS) * time.Millisecond
		} else {
			return 0, fmt.Errorf("OPT_CONNECTTIMEOUT_MS must be int")
		}
	} else if connectTimeout_, ok := options[OPT_CONNECTTIMEOUT]; ok {
		if connectTimeout, ok = connectTimeout_.(time.Duration); !ok {
			if connectTimeoutS, ok := connectTimeout_.(int); ok {
				connectTimeout = time.Duration(connectTimeoutS) * time.Second
			} else {
				return 0, fmt.Errorf("OPT_CONNECTTIMEOUT must be int or time.Duration")
			}
		}
	}

	timeout, err := prepareTimeout(options)
	if err != nil {
		return 0, err
	}

	// fix connect timeout(important, or it might cause a long time wait during
	//connection)
	if timeout > 0 && (connectTimeout > timeout || connectTimeout == 0) {
		connectTimeout = timeout
	}

	return connectTimeout, nil
}

// Prepare a transport.
//
// Handles timemout, proxy and maybe other transport related options here.
func prepareTransport(options map[int]interface{}) (http.RoundTripper, error) {
	transport := &http.Transport{}

	connectTimeout, err := prepareConnTimeout(options)
	if err != nil {
		return nil, err
	}

	transport.Dial = func(network, addr string) (net.Conn, error) {
		var conn net.Conn
		var err error
		if connectTimeout > 0 {
			conn, err = net.DialTimeout(network, addr, connectTimeout)
			if err != nil {
				return nil, err
			}
		} else {
			conn, err = net.Dial(network, addr)
			if err != nil {
				return nil, err
			}
		}

		return conn, nil
	}

	// proxy
	if proxyFunc_, ok := options[OPT_PROXY_FUNC]; ok {
		if proxyFunc, ok := proxyFunc_.(func(*http.Request) (int, string, error)); ok {
			transport.Proxy = func(req *http.Request) (*url.URL, error) {
				proxyType, u_, err := proxyFunc(req)
				if err != nil {
					return nil, err
				}

				if proxyType != PROXY_HTTP {
					return nil, fmt.Errorf("only PROXY_HTTP is currently supported")
				}

				u_ = "http://" + u_

				u, err := url.Parse(u_)

				if err != nil {
					return nil, err
				}

				return u, nil
			}
		} else {
			return nil, fmt.Errorf("OPT_PROXY_FUNC is not a desired function")
		}
	} else {
		var proxytype int
		if proxytype_, ok := options[OPT_PROXYTYPE]; ok {
			if proxytype, ok = proxytype_.(int); !ok || proxytype != PROXY_HTTP {
				return nil, fmt.Errorf("OPT_PROXYTYPE must be int, and only PROXY_HTTP is currently supported")
			}
		}

		var proxy string
		if proxy_, ok := options[OPT_PROXY]; ok {
			if proxy, ok = proxy_.(string); !ok {
				return nil, fmt.Errorf("OPT_PROXY must be string")
			}

			if !strings.Contains(proxy, "://") {
				proxy = "http://" + proxy
			}
			proxyUrl, err := url.Parse(proxy)
			if err != nil {
				return nil, err
			}
			transport.Proxy = http.ProxyURL(proxyUrl)
		}
	}

	// TLS
	if unsafe_tls_, found := options[OPT_UNSAFE_TLS]; found {
		var unsafe_tls, _ = unsafe_tls_.(bool)
		var tls_config = transport.TLSClientConfig
		if tls_config == nil {
			tls_config = &tls.Config{}
			transport.TLSClientConfig = tls_config
		}
		tls_config.InsecureSkipVerify = unsafe_tls
	}

	return transport, nil
}

// Prepare a redirect policy.
func prepareRedirect(options map[int]interface{}) (func(req *http.Request, via []*http.Request) error, error) {
	var redirectPolicy func(req *http.Request, via []*http.Request) error

	if redirectPolicy_, ok := options[OPT_REDIRECT_POLICY]; ok {
		if redirectPolicy, ok = redirectPolicy_.(func(*http.Request, []*http.Request) error); !ok {
			return nil, fmt.Errorf("OPT_REDIRECT_POLICY is not a desired function")
		}
	} else {
		var followlocation bool
		if followlocation_, ok := options[OPT_FOLLOWLOCATION]; ok {
			if followlocation, ok = followlocation_.(bool); !ok {
				return nil, fmt.Errorf("OPT_FOLLOWLOCATION must be bool")
			}
		}

		var maxredirs int
		if maxredirs_, ok := options[OPT_MAXREDIRS]; ok {
			if maxredirs, ok = maxredirs_.(int); !ok {
				return nil, fmt.Errorf("OPT_MAXREDIRS must be int")
			}
		}

		redirectPolicy = func(req *http.Request, via []*http.Request) error {
			// no follow
			if !followlocation || maxredirs <= 0 {
				return &Error{
					Code:    ERR_REDIRECT_POLICY,
					Message: fmt.Sprintf("redirect not allowed"),
				}
			}

			if len(via) >= maxredirs {
				return &Error{
					Code:    ERR_REDIRECT_POLICY,
					Message: fmt.Sprintf("stopped after %d redirects", len(via)),
				}
			}

			last := via[len(via)-1]
			// keep necessary headers
			// TODO: pass all headers or add other headers?
			if useragent := last.Header.Get("User-Agent"); useragent != "" {
				req.Header.Set("User-Agent", useragent)
			}

			return nil
		}
	}

	return redirectPolicy, nil
}

// Prepare a cookie jar.
func prepareJar(options map[int]interface{}) (http.CookieJar, error) {
	var jar http.CookieJar
	var err error
	if optCookieJar_, ok := options[OPT_COOKIEJAR]; ok {
		// is bool
		if optCookieJar, ok := optCookieJar_.(bool); ok {
			// default jar
			if optCookieJar {
				// TODO: PublicSuffixList
				jar, err = cookiejar.New(nil)
				if err != nil {
					return nil, err
				}
			}
		} else if optCookieJar, ok := optCookieJar_.(http.CookieJar); ok {
			jar = optCookieJar
		} else {
			return nil, fmt.Errorf("invalid cookiejar")
		}
	}

	return jar, nil
}

// Create an HTTP client.
func NewHttpClient() *HttpClient {
	c := &HttpClient{
		reuseTransport: true,
		reuseJar:       true,
		lock:           new(sync.Mutex),
	}

	return c
}

// Powerful and easy to use HTTP client.
type HttpClient struct {
	// Default options of this client.
	options map[int]interface{}

	// Default headers of this client.
	Headers map[string]string

	// Options of current request.
	oneTimeOptions map[int]interface{}

	// Headers of current request.
	oneTimeHeaders map[string]string

	// Cookies of current request.
	oneTimeCookies []*http.Cookie

	// Global transport of this client, might be shared between different
	// requests.
	transport http.RoundTripper

	// Global cookie jar of this client, might be shared between different
	// requests.
	jar http.CookieJar

	// Whether current request should reuse the transport or not.
	reuseTransport bool

	// Whether current request should reuse the cookie jar or not.
	reuseJar bool

	// Make requests of one client concurrent safe.
	lock *sync.Mutex

	withLock bool
}

// Set default options and headers.
func (this *HttpClient) Defaults(defaults Map) *HttpClient {
	options, headers := parseMap(defaults)

	// merge options
	if this.options == nil {
		this.options = options
	} else {
		for k, v := range options {
			this.options[k] = v
		}
	}

	// merge headers
	if this.Headers == nil {
		this.Headers = headers
	} else {
		for k, v := range headers {
			this.Headers[k] = v
		}
	}

	return this
}

// Begin marks the begining of a request, it's necessary for concurrent
// requests.
func (this *HttpClient) Begin() *HttpClient {
	this.lock.Lock()
	this.withLock = true

	return this
}

// Reset the client state so that other requests can begin.
func (this *HttpClient) reset() {
	this.oneTimeOptions = nil
	this.oneTimeHeaders = nil
	this.oneTimeCookies = nil
	this.reuseTransport = true
	this.reuseJar = true

	// nil means the Begin has not been called, asume requests are not
	// concurrent.
	if this.withLock {
		this.withLock = false
		this.lock.Unlock()
	}
}

// Temporarily specify an option of the current request.
func (this *HttpClient) WithOption(k int, v interface{}) *HttpClient {
	if this.oneTimeOptions == nil {
		this.oneTimeOptions = make(map[int]interface{})
	}
	this.oneTimeOptions[k] = v

	// Conditions we cann't reuse the transport.
	if hasOption(k, transportOptions) {
		this.reuseTransport = false
	}

	// Conditions we cann't reuse the cookie jar.
	if hasOption(k, jarOptions) {
		this.reuseJar = false
	}

	return this
}

// Temporarily specify multiple options of the current request.
func (this *HttpClient) WithOptions(m Map) *HttpClient {
	options, _ := parseMap(m)
	for k, v := range options {
		this.WithOption(k, v)
	}

	return this
}

// Temporarily specify a header of the current request.
func (this *HttpClient) WithHeader(k string, v string) *HttpClient {
	if this.oneTimeHeaders == nil {
		this.oneTimeHeaders = make(map[string]string)
	}
	this.oneTimeHeaders[k] = v

	return this
}

// Temporarily specify multiple headers of the current request.
func (this *HttpClient) WithHeaders(m map[string]string) *HttpClient {
	for k, v := range m {
		this.WithHeader(k, v)
	}

	return this
}

// Specify cookies of the current request.
func (this *HttpClient) WithCookie(cookies ...*http.Cookie) *HttpClient {
	this.oneTimeCookies = append(this.oneTimeCookies, cookies...)

	return this
}

// Start a request, and get the response.
//
// Usually we just need the Get and Post method.
func (this *HttpClient) Do(method string, url string, headers map[string]string,
	body io.Reader) (*Response, error) {
	options := mergeOptions(defaultOptions, this.options, this.oneTimeOptions)
	headers = mergeHeaders(this.Headers, headers, this.oneTimeHeaders)
	cookies := this.oneTimeCookies

	var transport http.RoundTripper
	var jar http.CookieJar
	var err error

	// transport
	if this.transport == nil || !this.reuseTransport {
		transport, err = prepareTransport(options)
		if err != nil {
			this.reset()
			return nil, err
		}

		if this.reuseTransport {
			this.transport = transport
		}
	} else {
		transport = this.transport
	}

	// jar
	if this.jar == nil || !this.reuseJar {
		jar, err = prepareJar(options)
		if err != nil {
			this.reset()
			return nil, err
		}

		if this.reuseJar {
			this.jar = jar
		}
	} else {
		jar = this.jar
	}

	// timeout
	timeout, err := prepareTimeout(options)
	if err != nil {
		this.reset()
		return nil, err
	}

	redirect, err := prepareRedirect(options)
	if err != nil {
		this.reset()
		return nil, err
	}

	req, err := prepareRequest(method, url, headers, body, options)
	if err != nil {
		this.reset()
		return nil, err
	}

	if debugEnabled, ok := options[OPT_DEBUG]; ok {
		if debugEnabled.(bool) {
			dump, err := httputil.DumpRequestOut(req, true)
			if err == nil {
				fmt.Printf("%s\n", dump)
			}
		}
	}

	if jar != nil {
		jar.SetCookies(req.URL, cookies)
	} else {
		for _, cookie := range cookies {
			req.AddCookie(cookie)
		}
	}

	if ctx, ok := options[OPT_CONTEXT]; ok {
		if c, ok := ctx.(context.Context); ok {
			req = req.WithContext(c)
		}
	}

	beforeReqFunc := options[OPT_BEFORE_REQUEST_FUNC]

	// release lock
	this.reset()

	c := &http.Client{
		Transport:     transport,
		CheckRedirect: redirect,
		Jar:           jar,
		Timeout:       timeout,
	}

	if beforeReqFunc != nil {
		if f, ok := beforeReqFunc.(func(c *http.Client, r *http.Request)); ok {
			f(c, req)
		}
	}

	res, err := c.Do(req)

	return &Response{res}, err
}

// The HEAD request
func (this *HttpClient) Head(url string) (*Response, error) {
	return this.Do("HEAD", url, nil, nil)
}

// The GET request
func (this *HttpClient) Get(url string, params ...interface{}) (*Response, error) {
	for _, p := range params {
		url = addParams(url, toUrlValues(p))
	}

	return this.Do("GET", url, nil, nil)
}

// The DELETE request
func (this *HttpClient) Delete(url string, params ...interface{}) (*Response, error) {
	for _, p := range params {
		url = addParams(url, toUrlValues(p))
	}

	return this.Do("DELETE", url, nil, nil)
}

// The POST request
//
// With multipart set to true, the request will be encoded as
// "multipart/form-data".
//
// If any of the params key starts with "@", it is considered as a form file
// (similar to CURL but different).
func (this *HttpClient) Post(url string, params interface{}) (*Response,
	error) {
	t := checkParamsType(params)
	if t == 2 {
		return this.Do("POST", url, nil, toReader(params))
	}

	paramsValues := toUrlValues(params)
	// Post with files should be sent as multipart.
	if checkParamFile(paramsValues) {
		return this.PostMultipart(url, params)
	}

	headers := make(map[string]string)
	headers["Content-Type"] = "application/x-www-form-urlencoded"
	body := strings.NewReader(paramsValues.Encode())

	return this.Do("POST", url, headers, body)
}

// Post with the request encoded as "multipart/form-data".
func (this *HttpClient) PostMultipart(url string, params interface{}) (
	*Response, error) {
	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)

	paramsValues := toUrlValues(params)
	// check files
	for k, v := range paramsValues {
		for _, vv := range v {
			// is file
			if k[0] == '@' {
				err := addFormFile(writer, k[1:], vv)
				if err != nil {
					this.reset()
					return nil, err
				}
			} else {
				writer.WriteField(k, vv)
			}
		}
	}
	headers := make(map[string]string)

	headers["Content-Type"] = writer.FormDataContentType()
	err := writer.Close()
	if err != nil {
		this.reset()
		return nil, err
	}

	return this.Do("POST", url, headers, body)
}

func (this *HttpClient) sendJson(method string, url string, data interface{}) (*Response, error) {
	headers := make(map[string]string)
	headers["Content-Type"] = "application/json"

	var body []byte
	switch t := data.(type) {
	case []byte:
		body = t
	case string:
		body = []byte(t)
	default:
		var err error
		body, err = json.Marshal(data)
		if err != nil {
			this.reset()
			return nil, err
		}
	}

	return this.Do(method, url, headers, bytes.NewReader(body))
}

func (this *HttpClient) PostJson(url string, data interface{}) (*Response, error) {
	return this.sendJson("POST", url, data)
}

// The PUT request
func (this *HttpClient) Put(url string, body io.Reader) (*Response, error) {
	return this.Do("PUT", url, nil, body)
}

// Put json data
func (this *HttpClient) PutJson(url string, data interface{}) (*Response, error) {
	return this.sendJson("PUT", url, data)
}

// Patch json data
func (this *HttpClient) PatchJson(url string, data interface{}) (*Response, error) {
	return this.sendJson("PATCH", url, data)
}

// The OPTIONS request
func (this *HttpClient) Options(url string, params ...map[string]string) (*Response, error) {
	for _, p := range params {
		url = addParams(url, toUrlValues(p))
	}

	return this.Do("OPTIONS", url, nil, nil)
}

// The CONNECT request
func (this *HttpClient) Connect(url string, params ...map[string]string) (*Response, error) {
	for _, p := range params {
		url = addParams(url, toUrlValues(p))
	}

	return this.Do("CONNECT", url, nil, nil)
}

// The TRACE request
func (this *HttpClient) Trace(url string, params ...map[string]string) (*Response, error) {
	for _, p := range params {
		url = addParams(url, toUrlValues(p))
	}

	return this.Do("TRACE", url, nil, nil)
}

// The PATCH request
func (this *HttpClient) Patch(url string, params ...map[string]string) (*Response, error) {
	for _, p := range params {
		url = addParams(url, toUrlValues(p))
	}

	return this.Do("PATCH", url, nil, nil)
}

// Get cookies of the client jar.
func (this *HttpClient) Cookies(url_ string) []*http.Cookie {
	if this.jar != nil {
		u, _ := url.Parse(url_)
		return this.jar.Cookies(u)
	}

	return nil
}

// Get cookie values(k-v map) of the client jar.
func (this *HttpClient) CookieValues(url_ string) map[string]string {
	m := make(map[string]string)

	for _, c := range this.Cookies(url_) {
		m[c.Name] = c.Value
	}

	return m
}

// Get cookie value of a specified cookie name.
func (this *HttpClient) CookieValue(url_ string, key string) string {
	for _, c := range this.Cookies(url_) {
		if c.Name == key {
			return c.Value
		}
	}

	return ""
}
