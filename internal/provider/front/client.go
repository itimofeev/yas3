package front

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"net/http"

	"github.com/go-playground/validator/v10"
)

type Config struct {
	BasePath string `validate:"required"`
}

type Client struct {
	httpClient http.Client
	cfg        Config
}

func New(cfg Config) (*Client, error) {
	err := validator.New().Struct(cfg)
	if err != nil {
		return nil, fmt.Errorf("config validation error: %w", err)
	}

	return &Client{
		httpClient: http.Client{Transport: http.DefaultTransport},
		cfg:        cfg,
	}, nil
}

func (c *Client) UploadFile(ctx context.Context, fileName string, content []byte) error {
	url := fmt.Sprintf("%s/api/v1/uploadFile/%s?fileSize=%d", c.cfg.BasePath, fileName, len(content))
	uploadReq, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewReader(content))
	if err != nil {
		return err
	}
	resp, err := c.httpClient.Do(uploadReq)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("response code not 200: %d", resp.StatusCode)
	}

	return nil
}

func (c *Client) GetFile(ctx context.Context, fileName string) ([]byte, error) {
	url := c.cfg.BasePath + "/api/v1/getFile/" + fileName
	uploadReq, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, err
	}

	resp, err := c.httpClient.Do(uploadReq)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("response code not 200: %d", resp.StatusCode)
	}

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	return respBody, nil
}
