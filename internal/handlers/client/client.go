package client

import (
	"bytes"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"io"
	"log"
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

// CreateAccessURL creates an access URL.
func (c *Client) CreateAccessURL(server entity.Server) (entity.AccessURL, error) {
	const op = "client.CreateAccessURL"
	defer func() {
		if r := recover(); r != nil {
			log.Printf("Recovered in %s: %v", op, r)
		}
	}()

	apiURL := createURL(server.IPAddr, server.Port, server.Key)
	log.Printf("%s: sending create request to %s", op, apiURL)
	resp, err := c.sendCreateRequest(apiURL)
	if err != nil {
		log.Printf("%s: %v", op, err)
		return entity.AccessURL{}, fmt.Errorf("%s: %w", op, err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusCreated {
		err = fmt.Errorf("status: %d", resp.StatusCode)
		log.Printf("%s: %v", op, err)
		return entity.AccessURL{}, fmt.Errorf("%s: %w", op, err)
	}

	accessURL, err := parseResponse(resp.Body, apiURL)
	if err != nil {
		log.Printf("%s: %v", op, err)
		return entity.AccessURL{}, fmt.Errorf("%s: %w", op, err)
	}

	log.Printf("%s: successfully created access URL: %s", op, accessURL.AccessKey)
	return accessURL, nil
}

// RemoveAccessURLs removes access URLs.
func (c *Client) RemoveAccessURLs(accessURLs []entity.AccessURL) error {
	const op = "client.RemoveAccessURLs"
	defer func() {
		if r := recover(); r != nil {
			log.Printf("Recovered in %s: %v", op, r)
		}
	}()

	for _, u := range accessURLs {
		log.Printf("%s: sending remove request for URL: %s", op, u.ID)
		resp, err := c.sendRemoveRequest(fmt.Sprintf("%s/%s", u.ApiURL, u.ID))
		if err != nil {
			log.Printf("%s: %v", op, err)
			return fmt.Errorf("%s: %w", op, err)
		}

		if resp.StatusCode != http.StatusNoContent {
			err = fmt.Errorf("status: %d", resp.StatusCode)
			log.Printf("%s: %v", op, err)
			return fmt.Errorf("%s: %w", op, err)
		}

		log.Printf("%s: successfully removed URL: %s", op, u.ID)
	}

	return nil
}

// sendRemoveRequest sends a DELETE request to the specified API URL.
func (c *Client) sendRemoveRequest(apiURL string) (*http.Response, error) {
	const op = "client.sendRemoveRequest"
	defer func() {
		if r := recover(); r != nil {
			log.Printf("Recovered in %s: %v", op, r)
		}
	}()

	req, err := http.NewRequest(http.MethodDelete, apiURL, nil)
	if err != nil {
		log.Printf("%s: %v", op, err)
		return nil, fmt.Errorf("%s: %w", op, err)
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.client.Do(req)
	if err != nil {
		log.Printf("%s: %v", op, err)
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	log.Printf("%s: received response with status code %d", op, resp.StatusCode)
	return resp, nil
}

// sendCreateRequest sends a POST request to the specified API URL with a generated request body.
func (c *Client) sendCreateRequest(apiURL string) (*http.Response, error) {
	const op = "client.sendCreateRequest"
	defer func() {
		if r := recover(); r != nil {
			log.Printf("Recovered in %s: %v", op, r)
		}
	}()

	body, err := createPostRequest()
	if err != nil {
		log.Printf("%s: %v", op, err)
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	data, err := json.Marshal(body)
	if err != nil {
		log.Printf("%s: %v", op, err)
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	req, err := http.NewRequest(http.MethodPost, apiURL, bytes.NewBuffer(data))
	if err != nil {
		log.Printf("%s: %v", op, err)
		return nil, fmt.Errorf("%s: %w", op, err)
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.client.Do(req)
	if err != nil {
		log.Printf("%s: %v", op, err)
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	log.Printf("%s: received response with status code %d", op, resp.StatusCode)
	return resp, nil
}

// parseResponse parses JSON body into response struct.
func parseResponse(body io.Reader, apiURL string) (entity.AccessURL, error) {
	const op = "client.parseResponse"
	defer func() {
		if r := recover(); r != nil {
			log.Printf("Recovered in %s: %v", op, r)
		}
	}()

	var createdURL model.Response
	if err := json.NewDecoder(body).Decode(&createdURL); err != nil {
		log.Printf("%s: %v", op, err)
		return entity.AccessURL{}, fmt.Errorf("%s: %w", op, err)
	}

	log.Printf("%s: successfully parsed response for URL: %s", op, apiURL)
	return converter.ToEntityFromClient(createdURL, apiURL), nil
}

// createURL creates URL for API requests.
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
			Bytes: 1024 * 1024 * 1024 * 100 * 1000 * 250,
		},
	}

	log.Printf("Generated post request: %+v", req)
	return req, nil
}

// generateRandomPassword generates a random password with fixed length.
func generateRandomPassword() string {
	return gofakeit.Password(true, true, true, true, false, 10)
}
