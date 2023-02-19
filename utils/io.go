package utils

import (
	"errors"
	"io"
	"net/http"
)

func ReadUrl(url string, headers http.Header) ([]byte, error) {
	client := http.Client{}
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, errors.New("could not create request: " + err.Error())
	}

	req.Header = headers
	res, err := client.Do(req)
	if err != nil {
		return nil, errors.New("could not send request: " + err.Error())
	}

	if res.Body != nil {
		defer res.Body.Close()
		body, readErr := io.ReadAll(res.Body)
		if readErr != nil {
			return nil, errors.New("could not read response body: " + err.Error())
		}

		if res.StatusCode != 200 {
			return nil, errors.New("invalid status " + res.Status + ": " + string(body))
		}

		return body, nil
	}

	return nil, errors.New("response body is nil")
}
