package main

import (
	"bytes"
	"context"
	"fmt"
	"net/http"
)

type ServiceCaller struct {
	client *http.Client
}

func (sc ServiceCaller) callService(ctx context.Context, endpoint string, signature string, body string) (int, error) {
	r := bytes.NewBuffer([]byte(body))
	req, err := http.NewRequest(http.MethodPost, endpoint, r)
	if err != nil {
		return 0, err
	}
	req = req.WithContext(ctx)
	req.Header.Set("Content-Type", "application-/json")
	req.Header.Set("X-vtypeio-Hmac-SHA256", signature)

	resp, err := sc.client.Do(req)
	if err != nil {
		return 0, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return 0, fmt.Errorf("unexpected status code %d", resp.StatusCode)
	}

	return resp.StatusCode, nil
}
