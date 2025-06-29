package client

import (
	"github.com/docker/docker/client"
)

// Client is a wrapper around the docker client.
type Client struct {
	wrapped *client.Client
}

func NewClient() (*Client, error) {
	wrapped, err := client.NewClientWithOpts(client.WithAPIVersionNegotiation())
	if err != nil {
		return nil, err
	}
	return &Client{wrapped: wrapped}, nil
}

// Unwrap returns the underlying client.Client
func (c *Client) Unwrap() *client.Client {
	return c.wrapped
}
