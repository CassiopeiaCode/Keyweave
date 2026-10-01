package transport

import (
	"context"
	"net"
	"net/http"
	"net/url"
	"time"
)

type Factory struct {
	DialTimeout time.Duration
}

func (f Factory) ClientForProxy(proxy string) (*http.Client, error) {
	tr := &http.Transport{
		DialContext: (&net.Dialer{
			Timeout:   f.DialTimeout,
			KeepAlive: 30 * time.Second,
		}).DialContext,
		ForceAttemptHTTP2: true,
	}

	if proxy == "" || proxy == "direct" || proxy == "none" {
		tr.Proxy = nil
	} else {
		u, err := url.Parse(proxy)
		if err != nil {
			return nil, err
		}

		// net/http handles HTTP proxy and CONNECT for HTTPS.
		// URL userinfo is used for Basic proxy authentication.
		tr.Proxy = http.ProxyURL(u)
	}

	return &http.Client{Transport: tr}, nil
}

func DoWithContext(
	ctx context.Context,
	c *http.Client,
	req *http.Request,
) (*http.Response, error) {
	return c.Do(req.WithContext(ctx))
}
