package httpclient

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"time"

	"github.com/avast/retry-go"

	"github.com/tarmalonchik/golibs/logger"
	"github.com/tarmalonchik/golibs/trace"
)

type Client struct {
	logLevel      logger.Level
	timeout       time.Duration
	retryAttempts uint
	retryDelay    time.Duration
	httpClient    *http.Client
	logger        *logger.Logger
	loggerSender  logger.Sender
}

func NewClient(opts ...Opt) *Client {
	c := &Client{
		logLevel:      logger.LevelError,
		timeout:       5 * time.Second,
		retryAttempts: 1,
		retryDelay:    100 * time.Millisecond,
	}

	for i := range opts {
		opts[i](c)
	}

	loggerOpts := make([]logger.Opt, 0, 2)
	loggerOpts = append(loggerOpts, logger.WithLevel(c.logLevel))
	if c.loggerSender != nil {
		loggerOpts = append(loggerOpts, logger.WithSender(c.loggerSender))
	}

	c.httpClient = &http.Client{
		Timeout: c.timeout,
		Transport: &loggingTransport{
			parent: http.DefaultTransport,
			logger: logger.NewLogger(loggerOpts...),
		},
	}

	return c
}

type Response struct {
	Body       []byte
	StatusCode int
}

func (c *Client) Do(ctx context.Context, method string, url *url.URL, opts ...RequestOptions[any]) (*Response, error) {
	req := withOptions(opts...)

	httpReq, err := req.buildRequest(ctx, url, method)
	if err != nil {
		return nil, trace.FuncNameWithError(err)
	}

	resp, err := c.do(ctx, httpReq)
	if err != nil {
		return nil, trace.FuncNameWithError(err)
	}
	defer func() { _ = resp.Body.Close() }()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, trace.FuncNameWithErrorMsg(err, "read response body")
	}

	return &Response{Body: body, StatusCode: resp.StatusCode}, nil
}

func (c *Client) do(ctx context.Context, req *http.Request) (*http.Response, error) {
	var (
		resp *http.Response
		err  error
	)

	err = retry.Do(
		func() error {
			resp, err = c.httpClient.Do(req)
			if err != nil {
				return err
			}

			if resp.StatusCode >= http.StatusInternalServerError {
				return fmt.Errorf("internal server error: %d", resp.StatusCode)
			}

			return nil
		},
		retry.Attempts(c.retryAttempts),
		retry.Delay(c.retryDelay),
		retry.RetryIf(func(err error) bool {
			select {
			case <-ctx.Done():
				return false
			default:
				return true
			}
		}),
	)
	if err != nil {
		return nil, trace.FuncNameWithError(err)
	}

	return resp, nil
}

func withOptions(opts ...RequestOptions[any]) *request[any] {
	r := request[any]{
		QueryParams: make(url.Values),
		Headers:     make(http.Header),
	}

	for i := range opts {
		opts[i](&r)
	}

	return &r
}
