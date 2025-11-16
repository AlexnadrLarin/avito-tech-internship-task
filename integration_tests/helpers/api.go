package helpers

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
)

func GetAPIURL() string {
	apiURL := os.Getenv("API_URL")
	if apiURL == "" {
		panic("API_URL environment variable is required")
	}
	return apiURL
}

func MakeRequest(method, endpoint string, body interface{}) (*http.Response, error) {
	url := fmt.Sprintf("%s/api/v1%s", GetAPIURL(), endpoint)

	var reqBody []byte
	var err error
	if body != nil {
		reqBody, err = json.Marshal(body)
		if err != nil {
			return nil, fmt.Errorf("failed to marshal request body: %w", err)
		}
	}

	req, err := http.NewRequest(method, url, bytes.NewBuffer(reqBody))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to make request: %w", err)
	}

	return resp, nil
}

func ParseResponse(resp *http.Response, target interface{}) error {
	defer resp.Body.Close()

	if resp.StatusCode >= 400 {
		var errResp map[string]interface{}
		if err := json.NewDecoder(resp.Body).Decode(&errResp); err == nil {
			return fmt.Errorf("API error: %v", errResp)
		}
		return fmt.Errorf("HTTP error: %d", resp.StatusCode)
	}

	if target != nil {
		return json.NewDecoder(resp.Body).Decode(target)
	}

	return nil
}
