package httpclient

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/url"
	"time"

	"github.com/tarmalonchik/golibs/logger"
	"github.com/tarmalonchik/golibs/trace"
)

type Opt func(client *Client)

func WithLogLevel(lvl logger.Level) Opt {
	return func(v *Client) {
		v.logLevel = lvl
	}
}

func WithLogTimeout(timeout time.Duration) Opt {
	return func(v *Client) {
		v.timeout = timeout
	}
}

func WithRetry(count uint) Opt {
	return func(v *Client) {
		if count <= 0 {
			return
		}
		v.retryAttempts = count
	}
}

func WithRetryDelay(delay time.Duration) Opt {
	return func(v *Client) {
		v.retryDelay = delay
	}
}

func WithLoggerSender(sender logger.Sender) Opt {
	return func(v *Client) {
		v.loggerSender = sender
	}
}

type RequestOptions[T any] func(client *request[any])

type request[T any] struct {
	Path        string
	Body        any
	Headers     http.Header
	QueryParams url.Values
}

func (r *request[T]) getBody() ([]byte, error) {
	if r.Body == nil {
		return nil, nil
	}

	body, err := json.Marshal(r.Body)
	if err != nil {
		return nil, trace.FuncNameWithErrorMsg(err, "marshal body")
	}

	return body, nil
}

func (r *request[T]) buildRequest(ctx context.Context, url *url.URL, method string) (*http.Request, error) {
	body, err := r.getBody()
	if err != nil {
		return nil, trace.FuncNameWithErrorMsg(err, "get request body")
	}

	req, err := http.NewRequestWithContext(ctx, method, url.String(), bytes.NewBuffer(body))
	if err != nil {
		return nil, trace.FuncNameWithErrorMsg(err, "building request")
	}

	for i, val := range r.Headers {
		for _, item := range val {
			req.Header.Add(i, item)
		}
	}

	req.URL.Path = "/"
	req.URL = req.URL.JoinPath(r.Path)

	q := req.URL.Query()

	for i := range r.QueryParams {
		q.Add(i, r.QueryParams.Get(i))
	}

	req.URL.RawQuery = q.Encode()

	return req, nil
}

func WithPath(path string) RequestOptions[any] {
	return func(r *request[any]) {
		r.Path = path
	}
}

func WithBody[T any](body *T) RequestOptions[any] {
	return func(r *request[any]) {
		r.Body = body
	}
}

func WithHeaders(headers http.Header) RequestOptions[any] {
	return func(r *request[any]) {
		r.Headers = headers
	}
}

func WithQueryParams(queryParams url.Values) RequestOptions[any] {
	return func(r *request[any]) {
		r.QueryParams = queryParams
	}
}
