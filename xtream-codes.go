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

// XtreamClient is the client used to communicate with a Xtream-Codes server.
type XtreamClient struct {
	Username  string
	Password  string
	BaseURL   string
	UserAgent string

	ServerInfo ServerInfo
	UserInfo   UserInfo

	HTTP    *http.Client
	Context context.Context

	streams map[int]Stream
}

// NewClient returns an initialized XtreamClient with the given values.
func NewClient(username, password, baseURL string) (*XtreamClient, error) {
	_, parseURLErr := url.Parse(baseURL)
	if parseURLErr != nil {
		return nil, fmt.Errorf("error parsing url: %s", parseURLErr.Error())
	}

	httpClient := &http.Client{
		Transport: &http.Transport{DisableKeepAlives: true},
		Timeout:   15 * time.Second,
	}

	client := &XtreamClient{
		Username:  username,
		Password:  password,
		BaseURL:   baseURL,
		UserAgent: defaultUserAgent,
		HTTP:      httpClient,
		Context:   context.Background(),
		streams:   make(map[int]Stream),
	}

	authData, authErr := client.sendRequest("", nil)
	if authErr != nil {
		return nil, fmt.Errorf("error sending authentication request: %s", authErr.Error())
	}

	a := &AuthenticationResponse{}
	if jsonErr := json.Unmarshal(authData, &a); jsonErr != nil {
		return nil, fmt.Errorf("error unmarshaling json: %s", jsonErr.Error())
	}

	client.ServerInfo = a.ServerInfo
	client.UserInfo = a.UserInfo

	return client, nil
}

// NewClientWithContext returns an initialized XtreamClient with the given values.
func NewClientWithContext(ctx context.Context, username, password, baseURL string) (*XtreamClient, error) {
	c, err := NewClient(username, password, baseURL)
	if err != nil {
		return nil, err
	}
	c.Context = ctx
	return c, nil
}

// NewClientWithUserAgent returns an initialized XtreamClient with the given values.
func NewClientWithUserAgent(ctx context.Context, username, password, baseURL, userAgent string) (*XtreamClient, error) {
	defaultUserAgent = userAgent
	c, err := NewClient(username, password, baseURL)
	if err != nil {
		return nil, err
	}
	c.UserAgent = userAgent
	c.Context = ctx
	return c, nil
}
