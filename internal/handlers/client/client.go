package client

import (
	"bytes"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"io"
	"math/rand"
	"net/http"
	"telegram-bot/internal/entity"
	"telegram-bot/internal/handlers/converter"
	"telegram-bot/internal/handlers/model"

	"github.com/brianvoe/gofakeit"
)

type Client struct {
	client *http.Client
}

func New() *Client {
	return &Client{
		client: &http.Client{
			Transport: &http.Transport{
				TLSClientConfig: &tls.Config{
					InsecureSkipVerify: true,
				},
			},
		},
	}
}

// CreateAccessURLs create access urls.
func (c *Client) CreateAccessURLs(servers []entity.Server) ([]entity.AccessURL, error) {
	const op = "api.handlers.CreateAccessKey"

	accessURLs := make([]entity.AccessURL, 0, len(servers))
	for _, server := range servers {
		apiURL := createURL(server.IPAddr, server.Port, server.Key)

		resp, err := c.sendCreateRequest(apiURL)
		if err != nil {
			return nil, fmt.Errorf("%s: %w", op, err)
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusCreated {
			return nil, fmt.Errorf("status: %d, %s: %w", resp.StatusCode, op, err)
		}

		accessURL, err := parseResponse(resp.Body, apiURL)
		if err != nil {
			return nil, fmt.Errorf("%s: %w", op, err)
		}

		accessURLs = append(accessURLs, accessURL)
	}

	return accessURLs, nil
}

// RemoveAccessURL remove access url.
func (c *Client) RemoveAccessURL(apiURL string, id string) error {
	const op = "api.handlers.Remove"

	resp, err := c.sendRemoveRequest(fmt.Sprintf("%s/%s", apiURL, id))
	if err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}

	if resp.StatusCode != http.StatusNoContent {
		return fmt.Errorf("status: %d, %s: %w", resp.StatusCode, op, err)
	}

	return nil
}

// sendRemoveRequest sends a DELETE request to the specified API URL.
func (c *Client) sendRemoveRequest(apiURL string) (*http.Response, error) {
	const op = "handlers.api.client.sendRemoveRequest"

	req, err := http.NewRequest(http.MethodDelete, apiURL, nil)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	return resp, nil
}

// sendCreateRequest sends a POST request to the specified API URL with a generated request body.
func (c *Client) sendCreateRequest(apiURL string) (*http.Response, error) {
	const op = "handlers.api.client.sendCreateRequest"

	body, err := createPostRequest()
	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	data, err := json.Marshal(body)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	req, err := http.NewRequest(http.MethodPost, apiURL, bytes.NewBuffer(data))
	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	return resp, nil
}

// parseResponse parse json body into response struct.
func parseResponse(body io.Reader, apiURL string) (entity.AccessURL, error) {
	const op = "hadnlers.api.client.parseReponse"

	var createdURL model.Response
	if err := json.NewDecoder(body).Decode(&createdURL); err != nil {
		return entity.AccessURL{}, fmt.Errorf("%s: %w", op, err)
	}

	return converter.ToEntityFromClient(createdURL, apiURL), nil
}

// createURL creates url for api requests.
func createURL(ipAddr string, port int, key string) string {
	return fmt.Sprintf("https://%s:%d/%s/access-keys", ipAddr, port, key)
}

// createPostRequest generates a new request body with random data.
func createPostRequest() (model.Request, error) {
	const method = "chacha20-ietf-poly1305"

	req := model.Request{
		Name:     gofakeit.Name(),
		Method:   method,
		Password: generateRandomPassword(),
		Port:     rand.Intn(60000),
		Limit: model.DataLimit{
			Bytes: 50000,
		},
	}

	return req, nil
}

// generateRandomPassword generate random password with fix length.
func generateRandomPassword() string {
	return gofakeit.Password(true, true, true, true, false, 10)
}
