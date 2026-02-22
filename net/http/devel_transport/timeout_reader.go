// Copyright 2025 the toolbox authors.
// SPDX-License-Identifier: Apache-2.0

package transport

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"net/http"
	"time"
)

// TODO: Use a buffer pool for storing the response body.

// TODO: Add the ability to set a maximum size for the download.
// This will help, since this is supposed to be for small files.

// TODO: Add a little comment justifying why this code exists
// and why does it have to be so darn complex.

// All the errors contain [ErrTimeout].
var (
	ErrTimeout        = fmt.Errorf("http request timed out")
	ErrConnectTimeout = fmt.Errorf("%w: connection attempt took too long", ErrTimeout)
	ErrReqIdle        = fmt.Errorf("%w: request body transmission was idle for too long", ErrTimeout)
	ErrReqTimeout     = fmt.Errorf("%w: request body transmission took too long", ErrTimeout)
	ErrRespIdle       = fmt.Errorf("%w: response body reception was idle for too long", ErrTimeout)
	ErrRespTimeout    = fmt.Errorf("%w: response body reception took too long", ErrTimeout)
)

// TimeoutReader is an [http.RoundTripper]
// that allows setting timeouts for different portions of the round trip,
// as well as reading the response body as part of the round trip
// so that it can be e.g. timed out separately
// or retried with the [Retrier] roundtripper.
type TimeoutReader struct {
	ConnectTimeout time.Duration
	ReqIdle        time.Duration
	ReqTimeout     time.Duration
	RespBody       time.Duration
	RespTimeout    time.Duration

	// Whether or not to read the boy in the round trip.
	ReadBody bool

	// The transport to perform the requests with.
	// Defaults to [http.DefaultTransport] if not set.
	Transport http.RoundTripper
}

func (t TimeoutReader) errToTimeout(err error) time.Duration {
	switch err {
	case ErrConnectTimeout:
		return t.ConnectTimeout
	case ErrReqIdle:
		return t.ReqIdle
	case ErrReqTimeout:
		return t.ReqTimeout
	case ErrRespIdle:
		return t.RespBody
	case ErrRespTimeout:
		return t.RespTimeout
	default:
		panic("unreachable")
	}
}

func isErrIdle(err error) bool {
	switch err {
	case ErrReqIdle,
		ErrRespIdle:
		return true
	default:
		return false
	}
}

func (t TimeoutReader) hasTimeout() bool {
	return t.ConnectTimeout == 0 &&
		t.ReqIdle == 0 &&
		t.ReqTimeout == 0 &&
		t.RespBody == 0 &&
		t.RespTimeout == 0
}

func (t TimeoutReader) wrapErr(err error) error {
	return fmt.Errorf("%w (> %s)", err, t.errToTimeout(err))
}

func (t TimeoutReader) RoundTrip(req *http.Request) (*http.Response, error) {

	tr := t.Transport
	if tr == nil {
		tr = http.DefaultTransport
	}

	const (
		evReq = iota
		evReqClose
		evResp
	)
	evChan := (chan int)(nil)
	sendEv := func(v int) {
		if evChan != nil {
			evChan <- v
		}
	}

	ctx, cancel := context.WithCancelCause(req.Context())

	if t.hasTimeout() {

		evChan = make(chan int)

		go func() {

			timer := newStoppedTimer()
			timerIdle := newStoppedTimer()
			defer func() {
				timer.Stop()
				timerIdle.Stop()
			}()

			var cancelCause error
			var cancelCauseIdle error

			reqOnce := false
			respOnce := false

			setTimeout := func(err error) {
				if isErrIdle(err) {
					cancelCauseIdle = t.wrapErr(err)
					timerIdle.Reset(t.errToTimeout(err))
				} else {
					cancelCause = t.wrapErr(err)
					timer.Reset(t.errToTimeout(err))
				}
			}

			if t.ConnectTimeout > 0 {
				setTimeout(ErrConnectTimeout)
			}

			for {
				select {

				case <-ctx.Done():
					return

				case <-timer.C:
					cancel(cancelCause)

				case <-timerIdle.C:
					cancel(cancelCauseIdle)

				case ev := <-evChan:
					switch ev {

					case evReq:
						timer.Stop()
						timerIdle.Stop()
						if t.ReqTimeout > 0 && !reqOnce {
							reqOnce = true
							setTimeout(ErrReqTimeout)
						}
						if t.ReqIdle > 0 {
							setTimeout(ErrReqIdle)
						}

					case evReqClose:
						timer.Stop()
						timerIdle.Stop()
						if t.ConnectTimeout > 0 {
							setTimeout(ErrConnectTimeout)
						}

					case evResp:
						timer.Stop()
						timerIdle.Stop()
						if t.RespTimeout > 0 && !respOnce {
							respOnce = true
							setTimeout(ErrRespTimeout)
						}
						if t.RespBody > 0 {
							setTimeout(ErrRespIdle)
						}
					}
				}
			}
		}()
	}

	body := req.Body
	hasBody := req.Body != nil && req.Body != http.NoBody
	if hasBody {
		r := readCloser{}
		body = r
		r.read = func(b []byte) (int, error) {
			sendEv(evReq)
			return req.Body.Read(b)
		}
		r.close = func() error {
			sendEv(evReqClose)
			return req.Body.Close()
		}
	}

	req = req.Clone(ctx)
	req.Body = body

	resp, err := tr.RoundTrip(req)
	if err != nil {
		cancel(nil)
		return resp, err
	}
	if !t.ReadBody {
		r := readCloser{}
		body := resp.Body
		resp.Body = r
		r.read = func(b []byte) (int, error) {
			return body.Read(b)
		}
		r.close = func() error {
			err := body.Close()
			cancel(nil)
			return err
		}
		return resp, err
	}

	defer func() {
		resp.Body.Close()
		cancel(nil)
	}()

	buf := make([]byte, 0, max(0, resp.ContentLength))
	w := writer{}
	w.write = func(b []byte) (int, error) {
		sendEv(evResp)
		buf = append(buf, b...)
		return len(b), nil
	}

	_, err = io.Copy(w, resp.Body)
	if err != nil {
		return resp, err
	}

	resp.Body = io.NopCloser(bytes.NewReader(buf))
	return resp, nil
}

func newStoppedTimer() *time.Timer {
	t := time.NewTimer(time.Hour)
	t.Stop()
	return t
}

type readCloser struct {
	read  func([]byte) (int, error)
	close func() error
}

func (r readCloser) Read(b []byte) (int, error) {
	return r.read(b)
}

func (r readCloser) Close() error {
	return r.close()
}

type writer struct {
	write func([]byte) (int, error)
}

func (w writer) Write(b []byte) (int, error) {
	return w.write(b)
}
