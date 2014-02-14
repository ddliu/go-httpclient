package httpclient

import (
    "fmt"

    "strings"
    "bytes"

    "time"

    "os"
    "io"

    "path/filepath"

    "net"
    "net/http"
    "net/url"

    "mime/multipart"
)

// https://github.com/bagder/curl/blob/169fedbdce93ecf14befb6e0e1ce6a2d480252a3/packages/OS400/curl.inc.in
const (
    VERSION = "0.1.0"
    USERAGENT = "go-httpclient v" + VERSION

    PROXY_HTTP = 0
    PROXY_SOCKS4 = 4
    PROXY_SOCKS5 = 5
    PROXY_SOCKS4A = 6

    // CURL like OPT
    OPT_AUTOREFERER = 58
    OPT_FOLLOWLOCATION = 52
    OPT_CONNECTTIMEOUT = 78
    OPT_CONNECTTIMEOUT_MS = 156
    OPT_MAXREDIRS = 68
    OPT_PROXYTYPE = 101
    OPT_TIMEOUT = 13
    OPT_TIMEOUT_MS = 155
    OPT_COOKIEJAR = 10082
    OPT_INTERFACE = 10062
    OPT_PROXY = 10004
    OPT_REFERER = 10016
    OPT_USERAGENT = 10018

    // Other OPT
    OPT_REDIRECT_POLICY = 100000
    OPT_PROXY_FUNC = 100001
)

var CONST = map[string]int {
    "OPT_AUTOREFERER": 58,
    "OPT_FOLLOWLOCATION": 52,
    "OPT_CONNECTTIMEOUT": 78,
    "OPT_CONNECTTIMEOUT_MS": 156,
    "OPT_MAXREDIRS": 68,
    "OPT_PROXYTYPE": 101,
    "OPT_TIMEOUT": 13,
    "OPT_TIMEOUT_MS": 155,
    "OPT_COOKIEJAR": 10082,
    "OPT_INTERFACE": 10062,
    "OPT_PROXY": 10004,
    "OPT_REFERER": 10016,
    "OPT_USERAGENT": 10018,

    "OPT_REDIRECT_POLICY": 100000,
    "OPT_PROXY_FUNC": 100001,
}

var defaultOptions = map[int]interface{} {
    OPT_FOLLOWLOCATION: true,
    OPT_MAXREDIRS: 10,
    OPT_AUTOREFERER: true,
    OPT_USERAGENT: USERAGENT,
}

func NewHttpClient(options map[int]interface{}) *HttpClient {
    return &HttpClient{
        Options: mergeOptions(defaultOptions, options),
    }
}

type HttpClient struct {
    Options map[int]interface{}
}

// Start a HTTP request, and returns the response
func (this *HttpClient) Request(method string, url_ string, headers map[string]string, body io.Reader, options map[int]interface{}) (*http.Response, error) {
    options = mergeOptions(this.Options, options)

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

    transport := &http.Transport{}

    connectTimeoutMS := 0

    if connectTimeoutMS_, ok := options[OPT_CONNECTTIMEOUT_MS]; ok {
        if connectTimeoutMS, ok = connectTimeoutMS_.(int); !ok {
            return nil, fmt.Errorf("OPT_CONNECTTIMEOUT_MS must be int")
        }
    } else if connectTimeout_, ok := options[OPT_CONNECTTIMEOUT]; ok {
        if connectTimeout, ok := connectTimeout_.(int); ok {
            connectTimeoutMS = connectTimeout * 1000    
        } else {
            return nil, fmt.Errorf("OPT_CONNECTTIMEOUT must be int")
        }
    }

    timeoutMS := 0

    if timeoutMS_, ok := options[OPT_TIMEOUT_MS]; ok {
        if timeoutMS, ok = timeoutMS_.(int); !ok {
            return nil, fmt.Errorf("OPT_TIMEOUT_MS must be int")
        }
    } else if timeout_, ok := options[OPT_TIMEOUT]; ok {
        if timeout, ok := timeout_.(int); ok {
            timeoutMS = timeout * 1000    
        } else {
            return nil, fmt.Errorf("OPT_TIMEOUT must be int")
        }
    }

    transport.Dial = func (network, addr string) (net.Conn, error) {
        conn, err := net.DialTimeout(network, addr, time.Duration(connectTimeoutMS) * time.Millisecond)
        if err != nil {
            return nil, err
        }

        if timeoutMS > 0 {
            conn.SetDeadline(time.Now().Add(time.Duration(timeoutMS) * time.Millisecond))
        }

        return conn, nil
    }

    // proxy
    if proxyFunc_, ok := options[OPT_PROXY_FUNC]; ok {
        if proxyFunc, ok := proxyFunc_.(func (*http.Request) (int, string, error)); ok {
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
            proxy = "http://" + proxy
            proxyUrl, err := url.Parse(proxy)
            if err != nil {
                return nil, err
            }
            transport.Proxy = http.ProxyURL(proxyUrl)
        }
    }

    // redirect
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
                return fmt.Errorf("redirect not allowed")
            }

            if len(via) >= maxredirs {
                return fmt.Errorf("stopped after %d redirects", len(via))
            }

            return nil
        }
    }

    client := &http.Client{
        Transport: transport,
        CheckRedirect: redirectPolicy,
    }
    return client.Do(req)
}

// The GET request
func (this *HttpClient) Get(url string, params map[string]string, headers map[string]string, options map[int]interface{}) (*http.Response, error) {
    url = addParams(url, params)

    return this.Request("GET", url, headers, nil, options)
}

// The POST request
// 
// With multipart set to true, the request will be encoded as "multipart/form-data". 
// If any of the params key starts with "@", it is considered as a form file (similar to CURL but different).
func (this *HttpClient) Post(url string, params map[string]string, headers map[string]string, isMultipart bool, options map[int]interface{}) (*http.Response, error) {
    var body io.Reader

    if isMultipart {
        body = &bytes.Buffer{}
        bodyWriter, _ := body.(io.Writer)
        writer := multipart.NewWriter(bodyWriter)

        // check files
        for k, v := range params {
            // is file
            if k[0] == '@' {
                err := addFormFile(writer, k[1:], v)
                if err != nil {
                    return nil, err
                }
            } else {
                writer.WriteField(k, v)
            }
        }
        if headers == nil {
            headers = make(map[string]string)
        }

        headers["Content-Type"] = writer.FormDataContentType()
        err := writer.Close()
        if err != nil {
            return nil, err
        }

    } else {
        if headers == nil {
            headers = make(map[string]string)
        }
        headers["Content-Type"] = "application/x-www-form-urlencoded"
        body = strings.NewReader(paramsToString(params))
    }

    return this.Request("POST", url, headers, body, options)
}

func paramsToString(params map[string]string) string {
    values := url.Values{}
    for k, v := range(params) {
        values.Set(k, v)
    }

    return values.Encode()
}

func addParams(url_ string, params map[string]string) string {
    if len(params) == 0 {
        return url_
    }

    if !strings.Contains(url_, "?") {
        url_ += "?"
    }

    if strings.HasSuffix(url_, "?") || strings.HasSuffix(url_, "&") {
        url_ += paramsToString(params)
    } else {
        url_ += "&" + paramsToString(params)
    }

    return url_
}

func addFormFile(writer *multipart.Writer, name, path string) error{
    file, err := os.Open(path)
    if err != nil {
        return err
    }
    defer file.Close()
    part, err := writer.CreateFormFile(name, filepath.Base(path))
    if err != nil {
        return err
    }

    _, err = io.Copy(part, file)

    return err
}

// Convert options with strin keys to desired format
func Option(o map[string]interface{}) map[int]interface{} {
    rst := make(map[int]interface{})
    for k, v := range o {
        k := "OPT_" + strings.ToUpper(k)
        if num, ok := CONST[k]; ok {
            rst[num] = v
        }
    }

    return rst
}

// Merge options(latter ones has higher priority)
func mergeOptions(options ...map[int]interface{}) map[int]interface{} {
    rst := make(map[int]interface{})

    for _, m := range options {
        for k, v := range m {
            rst[k] = v
        }
    }

    return rst
}