package httpclient

import (
    "time"
)

type Options {
    AutoReferer bool
    FollowLocation bool
    ConnectTimeout time.Duration
    Timeout time.Duration
    Headers map[string]string
}