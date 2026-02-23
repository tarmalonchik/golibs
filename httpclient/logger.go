package httpclient

import (
	"net/http"
	"net/http/httputil"

	"go.uber.org/zap"

	"github.com/tarmalonchik/golibs/logger"
)

type loggingTransport struct {
	parent http.RoundTripper
	logger *logger.Logger
}

func (s *loggingTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	dump, err := httputil.DumpRequest(req, true)
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
