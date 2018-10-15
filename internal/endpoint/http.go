package endpoint

import (
	"bytes"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"time"
)

const (
	httpExpiresAfter       = time.Second * 30
	httpRequestTimeout     = time.Second * 5
	httpMaxIdleConnections = 20
)

// HTTPConn is an endpoint connection
type HTTPConn struct {
	ep     Endpoint
	client *http.Client
}

func newHTTPConn(ep Endpoint) *HTTPConn {
	return &HTTPConn{
		ep: ep,
		client: &http.Client{
			Transport: &http.Transport{
				MaxIdleConnsPerHost: httpMaxIdleConnections,
				IdleConnTimeout:     httpExpiresAfter,
			},
			Timeout: httpRequestTimeout,
		},
	}
}

// Expired returns true if the connection has expired
func (conn *HTTPConn) Expired() bool {
	return false
}

// Send sends a message
func (conn *HTTPConn) Send(msg string) error {
	req, err := http.NewRequest("POST", conn.ep.Original, bytes.NewBufferString(msg))
	if err != nil {
		return err
	}

	req.Header.Set("Content-Type", "application/json")
	resp, err := conn.client.Do(req)
	if err != nil {
		return err
	}
	// close the connection to reuse it
	defer resp.Body.Close()
	// discard response
	if _, err := io.Copy(ioutil.Discard, resp.Body); err != nil {
		return err
	}
	// we only care about the 200 response
	if resp.StatusCode != 200 {
		return fmt.Errorf("invalid status: %s", resp.Status)
	}
	return nil
}
