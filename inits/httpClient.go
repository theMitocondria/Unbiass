package inits

import (
    "net/http"
    "time"
)

var (
    // HTTPClient is the shared HTTP client instance
    HTTPClient *http.Client
)

func InitHTTPClient() {
    HTTPClient = &http.Client{
        Timeout: time.Second * 180,
        Transport: &http.Transport{
            MaxIdleConns:        100,
            MaxIdleConnsPerHost: 100,
            IdleConnTimeout:     90 * time.Second,
        },
    }
}


