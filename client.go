package client

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

type Client struct {
	httpClient *http.Client
	BaseURL    string

	Settings *SettingsService
	Server   *ServerService
	Chat     *ChatService
}

type clientOpt func(c *Client) error

func NewClient(opts ...clientOpt) (*Client, error) {
	c := &Client{
		httpClient: http.DefaultClient,
		BaseURL:    "http://localhost:1865",
	}

	for _, opt := range opts {
		if err := opt(c); err != nil {
			return nil, err
		}
	}

	c.Settings = &SettingsService{c}
	c.Server = &ServerService{c}
	c.Chat = &ChatService{c}

	return c, nil
}

func WithHttpClient(httpClient *http.Client) clientOpt {
	return func(c *Client) error {
		c.httpClient = httpClient
		return nil
	}
}

func WithBaseURL(baseURL string) clientOpt {
	return func(c *Client) error {
		c.BaseURL = baseURL
		return nil
	}
}

type CatResponse[T any] struct {
	StatusCode int
	Value      T
	Raw        []byte
}

type CatServerError struct {
	StatusCode int
	Message    string
	Raw        []byte
}

func (s *CatServerError) Error() string {
	return fmt.Sprintf(
		"code: %d - msg: %s - %s",
		s.StatusCode, s.Message, s.Raw,
	)
}

func get[R any](ctx context.Context, c *Client, path string, response R) (*CatResponse[R], error) {
	return do(ctx, c, http.MethodGet, path, nil, response)
}

func post[R any](ctx context.Context, c *Client, path string, payload any, response R) (*CatResponse[R], error) {
	return do(ctx, c, http.MethodPost, path, payload, response)
}

func put[R any](ctx context.Context, c *Client, path string, payload any, response R) (*CatResponse[R], error) {
	return do(ctx, c, http.MethodPut, path, payload, response)
}

func del[R any](ctx context.Context, c *Client, path string, response R) (*CatResponse[R], error) {
	return do(ctx, c, http.MethodDelete, path, nil, response)
}

func do[R any](ctx context.Context, c *Client, method, path string, payload any, response R) (*CatResponse[R], error) {
	requestBody := new(bytes.Buffer)
	if payload != nil {
		err := json.NewEncoder(requestBody).Encode(payload)
		if err != nil {
			return nil, err
		}
	}

	url := c.BaseURL + path
	req, err := http.NewRequestWithContext(ctx, method, url, requestBody)
	if err != nil {
		return nil, err
	}

	res, err := c.httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()

	catResp := &CatResponse[R]{
		StatusCode: res.StatusCode,
	}

	responseBody, err := io.ReadAll(res.Body)
	if err != nil {
		return nil, err
	}

	if res.StatusCode > 399 {
		return nil, &CatServerError{
			StatusCode: res.StatusCode,
			Raw:        responseBody,
		}
	}

	err = json.Unmarshal(responseBody, &response)
	if err != nil {
		return nil, err
	}
	catResp.Value = response

	return catResp, nil
}
