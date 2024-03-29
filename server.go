package client

import (
	"context"
)

type ServerService struct {
	client *Client
}

type Version struct {
	Status  string `json:"status"`
	Version string `json:"version"`
}

func (v *ServerService) Version(ctx context.Context) (*Version, error) {
	resp, err := get(ctx, v.client, "/", &Version{})
	if err != nil {
		return nil, err
	}
	return resp.Value, nil
}
