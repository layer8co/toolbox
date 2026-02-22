// Copyright 2025 the toolbox authors.
// SPDX-License-Identifier: Apache-2.0

package transport

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"net/http"
	"regexp"
	"sync"
	"time"
)

// TODO: Use a buffer pool for storing the request body.

type Retrier struct {
	MaxTries   int
	RetryDelay time.Duration

	// This function is called for each request error
	// and the retrier gives up immediately if it returns true.
	// Defaults to [DefaultIsErrFatal] if not set.
	IsErrFatal func(err error) bool

	// The transport to perform the requests with.
	// Defaults to [http.DefaultTransport] if not set.
	Transport http.RoundTripper
}

var fatalErrRegexp = regexp.MustCompile(fmt.Sprintf(
	`(?i)^(%s|%s(%s))\b`,
	`unsupported protocol`,
	`(net/)?http:\s*`,
	`nil request|no host in request|invalid (header|trailer|method)`,
))

func DefaultIsErrFatal(err error) bool {
	return fatalErrRegexp.MatchString(err.Error())
}

func (t Retrier) RoundTrip(req *http.Request) (*http.Response, error) {

	tr := t.Transport
	if tr == nil {
		tr = http.DefaultTransport
	}
	isErrFatal := t.IsErrFatal
	if isErrFatal == nil {
		isErrFatal = DefaultIsErrFatal
	}

	// We close the body on context cancellation,
	// which will hopefully unblock blocking reads of the body.
	closeBody := sync.OnceFunc(func() {
		req.Body.Close()
	})
	stopAfterFunc := context.AfterFunc(req.Context(), func() {
		closeBody()
	})
	defer func() {
		stopAfterFunc()
		closeBody()
	}()

	body := req.Body
	bodyData := bytes.Buffer{}
	hasBody := req.Body != nil && req.Body != http.NoBody
	cloneReq := func() *http.Request {
		r := req.Clone(req.Context())
		r.Body = body
		return r
	}

	// Perform the request and save the original body data
	// while transmitting it.
	if hasBody {
		body = io.NopCloser(io.TeeReader(req.Body, &bodyData))
	}
	resp, err := tr.RoundTrip(cloneReq())
	if err == nil || isErrFatal(err) {
		// Request finished. Return the results.
		return resp, err
	}

	// Request failed to finish. Read and save the rest of the body data,
	// and repeat the request.

	if hasBody {
		io.Copy(&bodyData, req.Body)
		closeBody()

		r := bytes.NewReader(bodyData.Bytes())
		body = io.NopCloser(r)

		c := cloneReq
		cloneReq = func() *http.Request {
			r.Seek(0, io.SeekStart)
			return c()
		}
	}

	var lastErr error

	i := t.MaxTries
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
		time.Sleep(t.RetryDelay)
		resp, err := tr.RoundTrip(cloneReq())
		if err == nil || isErrFatal(err) {
			return resp, err
		}
		lastErr = err
	}

	return nil, fmt.Errorf(
		"could not complete the request after %d tries: %w",
		t.MaxTries, lastErr,
	)
}
