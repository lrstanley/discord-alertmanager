// Copyright (c) Liam Stanley <me@liamstanley.io>. All rights reserved. Use
// of this source code is governed by the MIT license that can be found in
// the LICENSE file.

package alertmanager

import (
	"fmt"
	"net/url"
	"strings"

	"github.com/go-openapi/runtime"
	httptransport "github.com/go-openapi/runtime/client"
	"github.com/lrstanley/discord-alertmanager/internal/models"
	alclient "github.com/prometheus/alertmanager/api/v2/client"
)

type Client struct {
	*alclient.AlertmanagerAPI

	config models.ConfigAlertmanager
}

func NewClient(config models.ConfigAlertmanager, debug bool) (*Client, error) {
	uri, err := url.Parse(config.URL)
	if err != nil {
		return nil, fmt.Errorf("failed to parse alertmanager url: %w", err)
	}

	alTransport := &alclient.TransportConfig{
		Host:     uri.Host,
		BasePath: uri.Path,
		Schemes:  []string{uri.Scheme},
	}

	if alTransport.BasePath == "" || alTransport.BasePath == "/" {
		alTransport.BasePath = alclient.DefaultBasePath
	}

	c := &Client{
		AlertmanagerAPI: alclient.NewHTTPClientWithConfig(nil, alTransport),
		config:          config,
	}

	return c, nil
}

func (c *Client) HandleAuth(op *runtime.ClientOperation) {
	if c.config.Username != "" && c.config.Password != "" {
		op.AuthInfo = httptransport.BasicAuth(c.config.Username, c.config.Password)
	}
}

func (c *Client) URL() string {
	return strings.TrimSuffix(c.config.URL, "/")
}

func (c *Client) SilenceURL(id string) string {
	return fmt.Sprintf("%s/#/silences/%s", c.URL(), id)
}
