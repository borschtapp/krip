package utils

import (
	"fmt"
	"io"
	"net/http"
	"net/url"
	"time"
)

func ReadUrl(url string, headers http.Header) ([]byte, *url.URL, error) {
	client := http.Client{
		Timeout: 30 * time.Second,
	}
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, nil, fmt.Errorf("could not create request: %w", err)
	}

	req.Header = headers
	res, err := client.Do(req)
	if err != nil {
		return nil, nil, fmt.Errorf("could not send request: %w", err)
	}
	defer res.Body.Close()

	body, readErr := io.ReadAll(res.Body)
	if readErr != nil {
		return nil, nil, fmt.Errorf("could not read response body: %w", readErr)
	}

	if res.StatusCode != 200 {
		return nil, nil, fmt.Errorf("invalid status %d %s: %s", res.StatusCode, res.Status, string(body))
	}

	return body, res.Request.URL, nil
}
