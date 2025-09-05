// Package xtreamcodes provides a Golang interface to the Xtream-Codes IPTV Server API.
package xtreamcodes

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strconv"
	"time"
)

var defaultUserAgent = "go.xstream-codes (Go-http-client/1.1)"

// limiter channel om max gelijktijdige requests te beperken (hier 5)
var concurrencyCh = make(chan struct{}, 5)

// XtreamClient is the client used to communicate with a Xtream-Codes server.
type XtreamClient struct {
	Username  string
	Password  string
	BaseURL   string
	UserAgent string

	ServerInfo ServerInfo
	UserInfo   UserInfo

	// Our HTTP client to communicate with Xtream
	HTTP    *http.Client
	Context context.Context

	// We store an internal map of Streams for use with GetStreamURL
	streams map[int]Stream
}

// ... (rest van je code blijft gelijk tot aan sendRequest)

func (c *XtreamClient) sendRequest(action string, parameters url.Values) ([]byte, error) {
	file := "player_api.php"
	if action == "xmltv.php" {
		file = action
		action = "" // otherwise rip xmltv.php url
	}
	url := fmt.Sprintf("%s/%s?username=%s&password=%s", c.BaseURL, file, c.Username, c.Password)
	if action != "" {
		url = fmt.Sprintf("%s&action=%s", url, action)
	}
	if parameters != nil {
		url = fmt.Sprintf("%s&%s", url, parameters.Encode())
	}

	request, httpErr := http.NewRequest("GET", url, nil)
	if httpErr != nil {
		return nil, httpErr
	}

	request.Header.Set("User-Agent", c.UserAgent)
	request = request.WithContext(c.Context)

	// limiter slot claimen
	concurrencyCh <- struct{}{}
	defer func() { <-concurrencyCh }()

	// retry mechanisme
	var response *http.Response
	var err error
	for i := 0; i < 3; i++ { // max 3 pogingen
		response, err = c.HTTP.Do(request)
		if err == nil && response.StatusCode < 500 {
			break
		}
		if response != nil {
			response.Body.Close()
		}
		time.Sleep(time.Duration(i+1) * time.Second) // backoff
	}
	if err != nil {
		return nil, fmt.Errorf("cannot reach server. %v", err)
	}

	if response.StatusCode > 399 {
		return nil, fmt.Errorf("status code was %d, expected 2XX-3XX", response.StatusCode)
	}

	buf := &bytes.Buffer{}
	if _, copyErr := io.Copy(buf, response.Body); copyErr != nil {
		return nil, copyErr
	}

	if closeErr := response.Body.Close(); closeErr != nil {
		return nil, fmt.Errorf("cannot read response. %v", closeErr)
	}

	return buf.Bytes(), nil
}
