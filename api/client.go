package api

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"radio-tui/models"
)

type Client struct {
	BaseURL string
}

func NewClient() *Client {
	return &Client{
		BaseURL: "https://de1.api.radio-browser.info/json/stations/byname/",
	}
}

func (c *Client) Search(query string) ([]models.Station, error) {
	if query == "" {
		return []models.Station{}, nil
	}

	resp, err := http.Get(c.BaseURL + url.PathEscape(query))
	if err != nil {
		return nil, fmt.Errorf("API request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	var stations []models.Station
	if err := json.NewDecoder(resp.Body).Decode(&stations); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return stations, nil
}
