// Copyright 2014-2015 Liu Dong <ddliuhb@gmail.com>.
// Licensed under the MIT license.

package httpclient

import (
	"testing"
)

func TestDefaultClient(t *testing.T) {
	res, err := Get("http://httpbin.org/get", nil)

	if err != nil {
		t.Error("get failed", err)
	}

	if res.StatusCode != 200 {
		t.Error("Status Code not 200")
	}
}
