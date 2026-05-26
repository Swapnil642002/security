package dnsfilter

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
)

// NextDNSClient provides methods to interact with the NextDNS API for category blocking.
type NextDNSClient struct {
	ProfileID string
	APIKey    string
}

// BlockCategory blocks or unblocks a category (e.g., "streaming", "social")
func (c *NextDNSClient) BlockCategory(category string, block bool) error {
	url := fmt.Sprintf("https://api.nextdns.io/profiles/%s", c.ProfileID)
	payload, _ := json.Marshal(map[string]map[string]bool{
		"categories": {category: block},
	})
	req, err := http.NewRequest("PATCH", url, bytes.NewBuffer(payload))
	if err != nil {
		return err
	}
	req.Header.Set("X-Api-Key", c.APIKey)
	req.Header.Set("Content-Type", "application/json")
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode != 200 {
		return fmt.Errorf("NextDNS API error: %s", resp.Status)
	}
	return nil
}

// Example usage (call from your handler or service):
// client := dnsfilter.NextDNSClient{ProfileID: "YOUR_PROFILE_ID", APIKey: "YOUR_API_KEY"}
// err := client.BlockCategory("streaming", true) // Block streaming
// err := client.BlockCategory("streaming", false) // Unblock streaming
