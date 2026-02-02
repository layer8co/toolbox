package transport

import (
	"bytes"
	"fmt"
	"io"
	"net/http"
	"time"
)

// TODO: On the first request attempt of the retrier,
// read and save the main request's body at the same time
// as it's sending it to upstream.
// On subsequent attempts, use the saved body.

// TODO: Use the bufpool package to store the body once it's ready.

type Retrier struct {
	MaxTries   int
	RetryDelay time.Duration
	Transport  http.RoundTripper // The upstream transport to send the requests to.
}

func (r Retrier) RoundTrip(req *http.Request) (*http.Response, error) {

	if r.Transport == nil {
		r.Transport = http.DefaultTransport
	}

	var bodyReader *bytes.Reader

	if req.Body != nil && req.Body != http.NoBody {
		b, err := io.ReadAll(req.Body)
		if err != nil {
			return nil, fmt.Errorf("could not read request body: %w", err)
		}
		bodyReader = bytes.NewReader(b)
	}

	var lastErr error

	i := r.MaxTries
	if i == 0 {
		i = -1
	}

	for {
		if i == 0 {
			break
		}
		if i > 0 {
			i--
		}

		// Clone the request
		newReq := req.Clone(req.Context())

		// Renew the request body
		if bodyReader != nil {
			bodyReader.Seek(0, io.SeekStart)
			newReq.Body = io.NopCloser(bodyReader)
		}

		// Make the request and get a response
		resp, err := r.Transport.RoundTrip(newReq)
		if err == nil {
			return resp, nil
		}

		lastErr = err
		time.Sleep(r.RetryDelay)
	}

	return nil, fmt.Errorf(
		"failed to make the request after %d tries: %w",
		r.MaxTries, lastErr,
	)
}
