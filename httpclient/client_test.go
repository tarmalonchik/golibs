package httpclient

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/url"
	"path"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"github.com/tarmalonchik/golibs/logger"
)

func TestOnlyPtr(t *testing.T) {
	type sk struct {
		some string
	}

	var s sk

	WithBody(&s) // should compile
}

type req struct {
	Some string `json:"some"`
}

func TestGetRequest(t *testing.T) {
	tCases := []tCase{
		{
			path:       nil,
			method:     http.MethodGet,
			statusCode: http.StatusOK,
			respBody:   []byte("hello"),
		},
		{
			path:       []string{"hello", "some"},
			method:     http.MethodGet,
			statusCode: http.StatusOK,
			respBody:   []byte("hello"),
		},
		{
			path:       nil,
			method:     http.MethodPost,
			statusCode: http.StatusOK,
			respBody:   []byte("ok"),
			reqBody: req{
				Some: "world",
			},
			headers: http.Header{
				"pusul": []string{"pusul"},
			},
			queryParams: url.Values{
				"foo": []string{"bar"},
			},
		},
		{
			path:            []string{"hello", "some"},
			method:          http.MethodGet,
			statusCode:      http.StatusInternalServerError,
			requiredRetries: 3,
			err:             true,
		},
	}

	for i := range tCases {
		caseRunner(t, tCases[i])
	}
}

type tCase struct {
	path            []string
	method          string
	statusCode      int
	respBody        []byte
	reqBody         any
	headers         http.Header
	queryParams     url.Values
	requiredRetries int
	err             bool
}

func caseRunner(t *testing.T, item tCase) {
	ch := make(chan struct{})
	defer close(ch)
	paths := path.Join(append([]string{"/"}, item.path...)...)

	port := mustRandomTCPPort()
	addressPort := fmt.Sprintf("127.0.0.1:%d", port)

	mux := http.NewServeMux()

	if item.requiredRetries <= 0 {
		item.requiredRetries = 1
	}

	wg := sync.WaitGroup{}
	wg.Add(item.requiredRetries)

	mux.HandleFunc(paths, func(w http.ResponseWriter, r *http.Request) {
		require.Equal(t, item.method, r.Method)

		for i, val := range item.headers {
			require.Equal(t, val, r.Header.Values(i))
		}

		if item.reqBody != nil {
			actualBody, err := io.ReadAll(r.Body)
			require.NoError(t, err)

			requiredBody, err := json.Marshal(item.reqBody)
			require.NoError(t, err)

			require.Equal(t, requiredBody, actualBody)
		}

		if item.queryParams != nil {
			require.Equal(t, item.queryParams, r.URL.Query())
		}

		w.WriteHeader(item.statusCode)

		_, err := w.Write(item.respBody)
		require.NoError(t, err)

		wg.Done()
	})

	srv := &http.Server{
		Addr:    addressPort,
		Handler: mux,
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	go func() {
		<-ctx.Done()
		ct, ca := context.WithTimeout(context.Background(), 3*time.Second)
		defer ca()

		require.NoError(t, srv.Shutdown(ct))
	}()

	go func() {
		err := srv.ListenAndServe()
		if err != nil && !errors.Is(err, http.ErrServerClosed) && !errors.Is(err, context.Canceled) {
			panic(err)
		}
		ch <- struct{}{}
	}()

	baseURL, err := url.Parse(fmt.Sprintf("http://%s", addressPort))
	require.NoError(t, err)

	cli := NewClient(WithLogLevel(logger.LevelInfo), WithRetry(3))

	opt := make([]RequestOptions[any], 0)
	if paths != "" {
		opt = append(opt, WithPath(paths))
	}
	if item.reqBody != nil {
		opt = append(opt, WithBody(&item.reqBody))
	}
	if item.headers != nil {
		opt = append(opt, WithHeaders(item.headers))
	}
	if item.queryParams != nil {
		opt = append(opt, WithQueryParams(item.queryParams))
	}

	resp, err := cli.Do(ctx, item.method, baseURL, opt...)
	if item.err {
		require.Error(t, err)
	} else {
		require.NoError(t, err)
		require.Equal(t, item.statusCode, resp.StatusCode)
		require.Equal(t, resp.Body, item.respBody)
	}

	wg.Wait()

	cancel()
	<-ch
}

func randomTCPPort() (port int, err error) {
	lis, err := net.Listen("tcp", "localhost:")
	if err != nil {
		return 0, fmt.Errorf("can't create listener: %w", err)
	}
	defer func() {
		if errClose := lis.Close(); errClose != nil {
			err = errors.Join(err, errClose)
		}
	}()

	tcpAddr, ok := lis.Addr().(*net.TCPAddr)
	if !ok {
		return 0, fmt.Errorf("failed to cast to TCPAddr: %w", err)
	}
	return tcpAddr.Port, nil
}

func mustRandomTCPPort() int {
	p, err := randomTCPPort()
	if err != nil {
		panic(fmt.Errorf("get random tcp port: %w", err))
	}
	return p
}
