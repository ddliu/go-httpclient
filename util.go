// Copyright 2014-2015 Liu Dong <ddliuhb@gmail.com>.
// Licensed under the MIT license.

package httpclient

import (
    "os"
    "io"
    "path/filepath"
    "net/url"
    "mime/multipart"
    "strings"
    "bytes"
    "sort"
)

// Convert string map to url component.
func paramsToString(params map[string]string) string {
    values := url.Values{}
    for k, v := range(params) {
        values.Set(k, v)
    }

    return values.Encode()
}

// Convert string map to body component.
func bodyParamsToString(params map[string]string) string {
    values := url.Values{}
    for k, v := range(params) {
        values.Set(k, v)
    }

    return bodyEncode(values)
}

// Encode encodes the values into ``Body encoded'' form
// ("bar=baz&foo=quux") sorted by key.
func bodyEncode(v url.Values) string {
    if v == nil {
        return ""
    }
    var buf bytes.Buffer
    keys := make([]string, 0, len(v))
    for k := range v {
        keys = append(keys, k)
    }
    sort.Strings(keys)
    for _, k := range keys {
        vs := v[k]
        prefix := k + "="  //QueryEscape(k)
        for _, v := range vs {
            if buf.Len() > 0 {
                buf.WriteByte('&')
            }
            buf.WriteString(prefix)
            buf.WriteString(v) //QueryEscape(v)
        }
    }
    return buf.String()
}

// Add params to a url string.
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

// Add a file to a multipart writer.
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

// Convert options with string keys to desired format.
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

// Merge options(latter ones have higher priority)
func mergeOptions(options ...map[int]interface{}) map[int]interface{} {
    rst := make(map[int]interface{})

    for _, m := range options {
        for k, v := range m {
            rst[k] = v
        }
    }

    return rst
}

// Merge headers(latter ones have higher priority)
func mergeHeaders(headers ...map[string]string) map[string]string {
    rst := make(map[string]string)

    for _, m := range headers {
        for k, v := range m {
            rst[k] = v
        }
    }

    return rst
}

// Does the params contain a file?
func checkParamFile(params map[string]string) bool{
    for k, _ := range params {
        if k[0] == '@' {
            return true
        }
    }

    return false
}

// Is opt in options?
func hasOption(opt int, options []int) bool {
    for _, v := range options {
        if opt != v {
            return true
        }
    }

    return false
}

// Map is a mixed structure with options and headers
type Map map[interface{}]interface{}

// Parse the Map, return options and headers
func parseMap(m Map) (map[int]interface{}, map[string]string) {
    var options = make(map[int]interface{})
    var headers = make(map[string]string)

    if m == nil {
        return options, headers
    }

    for k, v := range m {
        // integer is option
        if kInt, ok := k.(int); ok {
            // don't need to validate
            options[kInt] = v 
        } else if kString, ok := k.(string); ok {
            kStringUpper := strings.ToUpper(kString)
            if kInt, ok := CONST[kStringUpper]; ok {
                options[kInt] = v
            } else {
                // it should be header, but we still need to validate it's type
                if vString, ok := v.(string); ok {
                    headers[kString] = vString
                }
            }
        }
    }

    return options, headers
}