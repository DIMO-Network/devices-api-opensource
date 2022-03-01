package services

import (
	"fmt"
	"net/http"
)

type SmartcarWebhookClient struct {
	HTTPClient *http.Client
	WebhookID  string
}

func (c *SmartcarWebhookClient) Subscribe(vehicleID, accessToken string) error {
	url := fmt.Sprintf(smartcarWebhookURL, vehicleID, c.WebhookID)
	req, err := http.NewRequest("POST", url, nil)
	if err != nil {
		return fmt.Errorf("failed to construct webhook subscription request: %w", err)
	}
	req.Header.Set("Authorization", "Bearer "+accessToken)
	req.Header.Set("SC-Unit-System", "metric")

	resp, err := c.HTTPClient.Do(req)
	if err != nil {
		return fmt.Errorf("failure making webhook subscription request: %w", err)
	}
	defer resp.Body.Close() //nolint

	if resp.StatusCode >= 400 {
		return fmt.Errorf("webhook subscription request returned status code %d", resp.StatusCode)
	}
	return nil
}

func (c *SmartcarWebhookClient) Unsubscribe(vehicleID, accessToken string) error {
	url := fmt.Sprintf(smartcarWebhookURL, vehicleID, c.WebhookID)
	req, err := http.NewRequest("DELETE", url, nil)
	if err != nil {
		return fmt.Errorf("failed to construct webhook deletion request: %w", err)
	}
	req.Header.Set("Authorization", "Bearer "+accessToken)
	resp, err := c.HTTPClient.Do(req)
	if err != nil {
		return fmt.Errorf("failed sending webhook deletion request: %w", err)
	}
	if resp.StatusCode >= 400 {
		return fmt.Errorf("webhook deletion request returned status code %d", resp.StatusCode)
	}
	return nil
}
