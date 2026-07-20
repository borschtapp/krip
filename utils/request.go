package utils

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"time"

	"github.com/borschtapp/krip/model"
)

var defaultHeaders = http.Header{
	"Referer":    {"https://www.google.com/"},
	"User-Agent": {"Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/144.0.0.0 Safari/537.36"},
}

var defaultClient = &http.Client{Timeout: 30 * time.Second}

type RequestConfig struct {
	Method  string
	URL     string
	Body    io.Reader
	Headers http.Header
}

func ExecuteRequest(config RequestConfig, options model.RequestOptions) ([]byte, *url.URL, error) {
	ctx := options.Context
	if ctx == nil {
		ctx = context.Background()
	}

	client := options.HttpClient
	if client == nil {
		client = defaultClient
	}

	// Build final headers: defaults < config.Headers < options.Headers
	headers := mergeHeaders(defaultHeaders, options.Headers)
	if len(config.Headers) > 0 {
		headers = mergeHeaders(headers, config.Headers)
	}

	req, err := http.NewRequestWithContext(ctx, config.Method, config.URL, config.Body)
	if err != nil {
		return nil, nil, fmt.Errorf("could not create request: %w", err)
	}
	req.Header = headers

	res, err := client.Do(req)
	if err != nil {
		return nil, nil, fmt.Errorf("could not send request: %w", err)
	}
	defer func() { _ = res.Body.Close() }()

	body, err := io.ReadAll(io.LimitReader(res.Body, 10*1024*1024))
	if err != nil {
		return nil, nil, fmt.Errorf("could not read response body: %w", err)
	}

	if res.StatusCode/100 != 2 {
		errBody := string(body)
		if len(errBody) > 500 {
			errBody = errBody[:500] + "..."
		}
		return nil, nil, fmt.Errorf("invalid status %d %s: %s", res.StatusCode, res.Status, errBody)
	}

	return body, res.Request.URL, nil
}

func mergeHeaders(base, extra http.Header) http.Header {
	merged := make(http.Header)
	for k, v := range base {
		merged[k] = append([]string(nil), v...)
	}
	for k, v := range extra {
		merged[k] = append([]string(nil), v...)
	}
	return merged
}
