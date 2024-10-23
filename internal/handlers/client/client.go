package client

import (
	"bytes"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"github.com/brianvoe/gofakeit"
	"log/slog"
	"math/rand"
	"net/http"
	"telegram-bot/internal/domain/models"
	"telegram-bot/internal/errors"
	"telegram-bot/pkg/logger/sl"
)

type Client struct {
	log *slog.Logger

	client *http.Client
}

type Options func(*Client)

func WithLogger(log *slog.Logger) Options {
	return func(c *Client) {
		c.log = log
	}
}

func WithClient() Options {
	return func(c *Client) {
		c.client = &http.Client{
			Transport: &http.Transport{
				TLSClientConfig: &tls.Config{
					InsecureSkipVerify: true,
				},
			},
		}
	}
}

func New(opts ...Options) (*Client, error) {
	const op = "client.New"

	c := new(Client)

	for _, opt := range opts {
		opt(c)
	}

	if c.client == nil {
		return nil, fmt.Errorf("%s: %w", op, errors.ErrNoClientProvided)
	}

	if c.log == nil {
		return nil, fmt.Errorf("%s: %w", op, errors.ErrNoLoggerProvided)
	}

	return c, nil
}

func (c *Client) CreateAccessKey(server models.Server) (models.AccessKey, error) {
	const op = "client.CreateAccessKey"

	log := c.log.With(slog.String("op", op))

	url := createAccessKeyURL(server.IpAddress, server.Port, server.Key)

	log.Info("attempting to send post request")

	requestBody, err := createPostRequestBody()
	if err != nil {
		log.Error("failed to create request body", sl.Err(err))

		return models.AccessKey{}, fmt.Errorf("%s: %w", op, err)
	}

	data, err := json.Marshal(requestBody)
	if err != nil {
		log.Error("failed to marshal request body", sl.Err(err))

		return models.AccessKey{}, fmt.Errorf("%s: %w", op, err)
	}

	req, err := http.NewRequest(http.MethodPost, url, bytes.NewBuffer(data))
	if err != nil {
		log.Error("failed to create request", sl.Err(err))

		return models.AccessKey{}, fmt.Errorf("%s: %w", op, err)
	}

	req.Header.Set("Content-Type", "application/json")

	resp, err := c.client.Do(req)
	if err != nil {
		log.Error("failed to send request", sl.Err(err))

		return models.AccessKey{}, fmt.Errorf("%s: %w", op, err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusCreated {
		log.Error("failed to create access key", sl.Err(err))

		return models.AccessKey{}, fmt.Errorf("%s: %w", op, err)
	}

	var accessKey models.AccessKey
	if err := json.NewDecoder(resp.Body).Decode(&accessKey); err != nil {
		log.Error("failed to decode response body", sl.Err(err))

		return models.AccessKey{}, fmt.Errorf("%s: %w", op, err)
	}

	log.Info("access key created successfully", slog.String("access_key", accessKey.Key))
	return accessKey, nil
}

func (c *Client) DeleteAccessKeys(accessKeys []models.AccessKey) error {
	const op = "client.DeleteAccessKeys"

	log := c.log.With(slog.String("op", op))

	for _, accessKey := range accessKeys {
		log.Info("attempting to delete access key", slog.String("access_key", accessKey.Key))

		req, err := http.NewRequest(http.MethodDelete, accessKey.ApiURL, nil)
		if err != nil {
			log.Error("failed to create request", sl.Err(err))

			return fmt.Errorf("%s: %w", op, err)
		}
		req.Header.Set("Content-Type", "application/json")

		resp, err := c.client.Do(req)
		if err != nil {
			log.Error("failed to send request", sl.Err(err))

			return fmt.Errorf("%s: %w", op, err)
		}

		if resp.StatusCode != http.StatusNoContent {
			log.Error("failed to delete access key", sl.Err(err))

			return fmt.Errorf("%s: %w", op, err)
		}

		log.Info("access key deleted successfully", slog.String("access_key", accessKey.Key))
	}

	return nil
}

func createPostRequestBody() (models.Request, error) {
	const method = "chacha20-ietf-poly1305"

	return models.Request{
		Name:     gofakeit.Name(),
		Method:   method,
		Password: generateRandomPassword(),
		Port:     rand.Intn(60000),
		Limit: models.DataLimit{
			Bytes: 1024 * 1024 * 1024 * 1024 * 1024,
		},
	}, nil
}

func createAccessKeyURL(ipAddress string, port int, key string) string {
	return fmt.Sprintf("https://%s:%d/%s/access-keys", ipAddress, port, key)
}

func generateRandomPassword() string {
	return gofakeit.Password(true, true, true, true, false, 10)
}
