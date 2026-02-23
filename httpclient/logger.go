package httpclient

import (
	"bytes"
	"io"
	"net/http"
	"net/http/httputil"
	"strings"

	"go.uber.org/zap"

	"github.com/tarmalonchik/golibs/logger"
)

type loggingTransport struct {
	parent      http.RoundTripper
	logger      *logger.Logger
	maskHeaders map[string]struct{}
}

func (s *loggingTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	cloneReq, err := cloneRequestWithBody(req)
	if err != nil {
		return nil, err
	}
	cloneReq.Header = s.mask(cloneReq.Header)

	dump, err := httputil.DumpRequest(cloneReq, true)
	if err != nil {
		s.logger.Error("failed to dump http request", zap.Error(err))
	} else {
		s.logger.Info("http request dump", zap.String("dump", string(dump)))
	}

	resp, err := s.parent.RoundTrip(req)
	if err != nil {
		s.logger.Error("failed to do http request", zap.Error(err), zap.String("dump", string(dump)))
		return nil, err
	}

	dump, err = httputil.DumpResponse(resp, true)
	if err != nil {
		s.logger.Error("failed to dump http response", zap.Error(err))
		return resp, nil
	}

	s.logger.Info("http response dump", zap.String("dump", string(dump)))

	return resp, nil
}

func (s *loggingTransport) mask(header http.Header) http.Header {
	for key := range header {
		if _, ok := s.maskHeaders[strings.ToLower(key)]; ok {
			header.Set(key, "[masked]")
		}
	}
	return header
}

func cloneRequestWithBody(req *http.Request) (*http.Request, error) {
	newReq := req.Clone(req.Context())

	if req.Body == nil {
		return newReq, nil
	}

	bodyBytes, err := io.ReadAll(req.Body)
	if err != nil {
		return nil, err
	}

	defer func() {
		_ = req.Body.Close()
	}()

	req.Body = io.NopCloser(bytes.NewBuffer(bodyBytes))

	newReq.Body = io.NopCloser(bytes.NewBuffer(bodyBytes))
	return newReq, nil
}
