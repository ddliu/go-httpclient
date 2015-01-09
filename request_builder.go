package httpclient

import (
    "net/http"
)

func NewBuilder() *Builder {
    return &Builder {
        Header http.Header{}
    }
}

type Builder {
    Method string
    URL *url.URL
    Header http.Header

    File string
    FormFiles map[string]string
    Data map[string]string

}

func (this *Builder) Build(method string, url string) *http.Request {

}

func (this *Builder) SetFile(path string) {

}

func (this *Builder) AddFormFile(name string, path string) {

}

func AddField

func (this *Builder) AddHeader(k, v string) {
    this.Header.Set(k, v)
}

func Data()