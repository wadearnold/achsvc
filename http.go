// Copyright 2018 The Paygate Authors
// Use of this source code is governed by an Apache License
// license that can be found in the LICENSE file.

package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"time"

	"github.com/go-kit/kit/log"
	"github.com/go-kit/kit/metrics"
	"github.com/go-kit/kit/metrics/prometheus"
	"github.com/gorilla/mux"
)

const (
	// maxReadBytes is the number of bytes to read
	// from a request body. It's intended to be used
	// with an io.LimitReader
	maxReadBytes = 1 * 1024 * 1024
)

var (
	errNoUserId = errors.New("no X-User-Id header provided")
)

// read consumes an io.Reader (wrapping with io.LimitReader)
// and returns either the resulting bytes or a non-nil error.
func read(r io.Reader) ([]byte, error) {
	r = io.LimitReader(r, maxReadBytes)
	return ioutil.ReadAll(r)
}

// getUserId grabs the userId from the http header, which is
// trusted. (The infra ensures this)
func getUserId(r *http.Request) string {
	return r.Header.Get("X-User-Id")
}

// getIdempotencyKey extracts X-Idempotency-Key from the http request,
// which is used to ensure transactions only commit once.
func getIdempotencyKey(r *http.Request) string {
	if v := r.Header.Get("X-Idempotency-Key"); v != "" {
		return v
	}
	return nextID()
}

// getRequestId extracts X-Request-Id from the http request, which
// is used in tracing requests.
//
// TODO(adam): IIRC a "max header size" param in net/http.Server - verify and configure
func getRequestId(r *http.Request) string {
	return r.Header.Get("X-Request-Id")
}

// encodeError JSON encodes the supplied error
//
// The HTTP status of "400 Bad Request" is written to the
// response.
func encodeError(w http.ResponseWriter, err error) {
	if err == nil {
		return
	}
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(http.StatusBadRequest)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"error": err.Error(),
	})
}

func internalError(w http.ResponseWriter, err error, component string) {
	internalServerErrors.Add(1)
	logger.Log(component, err)
	w.WriteHeader(http.StatusInternalServerError)
}

func addPingRoute(r *mux.Router) {
	r.Methods("GET").Path("/ping").HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/plain")

		userId := getUserId(r)
		if userId == "" {
			w.Write([]byte("PONG"))
		} else {
			w.Write([]byte(fmt.Sprintf("hello %s", userId)))
		}
	})
}

// wrapResponseWriter
//
// labels is the kv pairs passed to the prometheus metric, it is used for logging.
func wrapResponseWriter(w http.ResponseWriter, m *prometheus.Histogram, labels []string) *paygateResponseWriter {
	return &paygateResponseWriter{
		ResponseWriter: w,
		start:          time.Now(),
		labels:         labels,
		metric:         m.With(labels...),
		log:            logger,
	}
}

// paygateResponseWriter
//
// This is not a thread-safe struct!
type paygateResponseWriter struct {
	http.ResponseWriter

	start  time.Time
	labels []string
	metric metrics.Histogram

	headersWritten    bool   // has .WriteHeader been called yet?
	userId, requestId string // X-Request-Id

	log log.Logger
}

// ensureHeaders verifies the headers which paygate cares about.
//  X-User-Id, X-Request-Id, and X-Idempotency-Key
//
// X-User-Id is required, and requests without one will be completed
// with a 403 forbidden.
//
// X-Request-Id is optional, but if used we will emit a log line with
// that request fulfilment timing and the status code.
//
// X-Idempotency-Key is optional, but recommended to ensure requests
// only execute once. Clients are assumed to resend requests many times
// with the same key. We just need to reply back "already done".
func (w *paygateResponseWriter) ensureHeaders(r *http.Request) error {
	if v := getUserId(r); v == "" {
		w.Header().Set("Content-Type", "text/plain")
		w.WriteHeader(http.StatusForbidden)
		return errNoUserId
	} else {
		w.userId = v
	}
	w.requestId = getRequestId(r)

	// TODO(adam): idempotency check with an inmem bloom filter?
	// https://github.com/steakknife/bloomfilter

	return nil
}

// WriteHeader intercepts our usual
func (w *paygateResponseWriter) WriteHeader(code int) {
	if w.headersWritten {
		return
	}
	w.headersWritten = true

	diff := time.Now().Sub(w.start)

	w.metric.Observe(diff.Seconds())

	if len(w.labels) > 1 && w.requestId != "" {
		line := fmt.Sprintf("status=%d, took=%s, userId=%s, requestId=%s", code, diff, w.userId, w.requestId)
		w.log.Log(w.labels[1], line) // labels has the form: []string{"route", "getUserEvents"}
	}

	w.ResponseWriter.WriteHeader(code)
}
