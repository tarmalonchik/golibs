package tgfile

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"mime/multipart"
	"net/http"
	"net/url"
	"strconv"

	"github.com/tarmalonchik/golibs/httpclient"
	"github.com/tarmalonchik/golibs/logger"
	"github.com/tarmalonchik/golibs/trace"
)

type Client interface {
	SendFileBytes(ctx context.Context, chatID int64, file []byte, fileName string) (*string, error)
	SendFileByID(ctx context.Context, chatID int64, fileID string) error
}

type client struct {
	config     Config
	logger     *logger.Logger
	httpClient httpclient.Client
}

func NewClient(httpClient httpclient.Client, config Config, l *logger.Logger) Client {
	if l == nil {
		l = logger.NewLogger()
	}

	return &client{
		config:     config,
		logger:     l,
		httpClient: httpClient,
	}
}

func (c *client) SendFileBytes(ctx context.Context, chatID int64, file []byte, fileName string) (*string, error) {
	const (
		method        = "/sendDocument"
		requestMethod = http.MethodPost
	)

	u, err := url.Parse(c.config.TgBotBaseURL)
	if err != nil {
		return nil, trace.FuncNameWithErrorMsg(err, "parsing base url")
	}
	u.Path += fmt.Sprintf(tokenTemp, c.config.TgBotToken) + method

	body := new(bytes.Buffer)
	writer := multipart.NewWriter(body)
	defer func() { _ = writer.Close() }()

	part, err := writer.CreateFormFile(documentField, fileName)
	if err != nil {
		return nil, trace.FuncNameWithErrorMsg(err, "create form file")
	}
	if _, err := part.Write(file); err != nil {
		return nil, trace.FuncNameWithErrorMsg(err, "write file")
	}

	if err = writer.WriteField(chatIDField, strconv.Itoa(int(chatID))); err != nil {
		return nil, trace.FuncNameWithErrorMsg(err, "write chat id")
	}

	if err = writer.Close(); err != nil {
		return nil, trace.FuncNameWithErrorMsg(err, "close writer")
	}

	httpReq, err := http.NewRequestWithContext(ctx, requestMethod, u.String(), body)
	if err != nil {
		return nil, trace.FuncNameWithErrorMsg(err, "create request")
	}
	httpReq.Header.Set("Content-Type", writer.FormDataContentType())

	resp, err := c.httpClient.DoRequest(ctx, httpReq)
	if err != nil {
		return nil, trace.FuncNameWithErrorMsg(err, "sending request")
	}

	if resp.StatusCode >= 400 {
		return nil, trace.FuncNameWithErrorMsg(errors.New("bad status code"), "bad status code")
	}

	fileID, err := extractFileIDData(resp.Body)
	if err != nil {
		return nil, trace.FuncNameWithErrorMsg(err, "extract file id")
	}

	return fileID, nil
}

func (c *client) SendFileByID(ctx context.Context, chatID int64, fileID string) error {
	const (
		method        = "/sendDocument"
		requestMethod = http.MethodPost
	)

	u, err := url.Parse(c.config.TgBotBaseURL)
	if err != nil {
		return trace.FuncNameWithErrorMsg(err, "parsing base url")
	}
	u.Path += fmt.Sprintf(tokenTemp, c.config.TgBotToken) + method

	body := new(bytes.Buffer)
	writer := multipart.NewWriter(body)
	defer func() { _ = writer.Close() }()

	if err = writer.WriteField(documentField, fileID); err != nil {
		return err
	}

	if err = writer.WriteField(chatIDField, strconv.Itoa(int(chatID))); err != nil {
		return trace.FuncNameWithErrorMsg(err, "write chat id")
	}

	if err = writer.Close(); err != nil {
		return trace.FuncNameWithErrorMsg(err, "close writer")
	}

	httpReq, err := http.NewRequestWithContext(ctx, requestMethod, u.String(), body)
	if err != nil {
		return trace.FuncNameWithErrorMsg(err, "create request")
	}
	httpReq.Header.Set("Content-Type", writer.FormDataContentType())

	resp, err := c.httpClient.DoRequest(ctx, httpReq)
	if err != nil {
		return trace.FuncNameWithErrorMsg(err, "sending request")
	}

	if resp.StatusCode >= 400 {
		return trace.FuncNameWithErrorMsg(errors.New("bad status code"), "bad status code")
	}

	return nil
}

func extractFileIDData(body []byte) (*string, error) {
	var responseData sendDocResponseBody

	if err := json.Unmarshal(body, &responseData); err != nil {
		return nil, trace.FuncNameWithErrorMsg(err, "unmarshal json")
	}

	return &responseData.Result.Document.FileID, nil
}
