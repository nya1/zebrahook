package main

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"sync"
	"time"
	frontsvr "zebrahook/gen/http/zebrahook/server"
	front "zebrahook/gen/zebrahook"

	"github.com/rs/zerolog"
	goahttp "goa.design/goa/v3/http"
	httpmdlwr "goa.design/goa/v3/http/middleware"
	"goa.design/goa/v3/middleware"
)

// from goa middleware
// shortID produces a " unique" 6 bytes long string.
// Do not use as a reliable way to get unique IDs, instead use for things like logging.
func shortID() string {
	b := make([]byte, 6)
	io.ReadFull(rand.Reader, b)
	return base64.RawURLEncoding.EncodeToString(b)
}

// from goa middleware, edited to work with zerolog instead of generic logger
// Log returns a middleware that logs incoming HTTP requests and outgoing
// responses. The middleware uses the request ID set by the RequestID middleware
// or creates a short unique request ID if missing for each incoming request and
// logs it with the request and corresponding response details.
//
// The middleware logs the incoming requests HTTP method and path as well as the
// originator of the request. The originator is computed by looking at the
// X-Forwarded-For HTTP header or - absent of that - the originating IP. The
// middleware also logs the response HTTP status code, body length (in bytes) and
// timing information.
func customZeroLog(l *zerolog.Logger) func(h http.Handler) http.Handler {
	return func(h http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// get request id, if not present generate new one
			reqID := r.Context().Value(middleware.RequestIDKey)
			if reqID == nil {
				reqID = shortID()
			}

			loggerWithRequestId := l.With().Str("requestId", reqID.(string)).Logger()
			*l = loggerWithRequestId

			started := time.Now()

			rw := httpmdlwr.CaptureResponse(w)
			h.ServeHTTP(rw, r)

			l.Info().Int64("responseTimeMs", time.Since(started).Milliseconds()).Int("contentLength", rw.ContentLength).Int("statusCode", rw.StatusCode).Msg(r.Method + " " + r.URL.String())
		})
	}
}

// handleHTTPServer starts configures and starts a HTTP server on the given
// URL. It shuts down the server if any error is received in the error channel.
func handleHTTPServer(ctx context.Context, u *url.URL, frontEndpoints *front.Endpoints, wg *sync.WaitGroup, errc chan error, logger *zerolog.Logger, debug bool) {

	// Provide the transport specific request decoder and response encoder.
	// The goa http package has built-in support for JSON, XML and gob.
	// Other encodings can be used by providing the corresponding functions,
	// see goa.design/implement/encoding.
	var (
		dec = goahttp.RequestDecoder
		enc = goahttp.ResponseEncoder
	)

	// Build the service HTTP request multiplexer and configure it to serve
	// HTTP requests to the service endpoints.
	var mux goahttp.Muxer
	{
		mux = goahttp.NewMuxer()
	}

	// Wrap the endpoints with the transport specific layers. The generated
	// server packages contains code generated from the design which maps
	// the service input and output data structures to HTTP requests and
	// responses.
	var (
		frontServer *frontsvr.Server
	)
	{
		eh := errorHandler(logger)
		frontServer = frontsvr.New(frontEndpoints, mux, dec, enc, eh, nil)
		if debug {
			servers := goahttp.Servers{
				frontServer,
			}
			servers.Use(httpmdlwr.Debug(mux, os.Stdout))
		}
	}
	// Configure the mux.
	frontsvr.Mount(mux, frontServer)

	// Wrap the multiplexer with additional middlewares. Middlewares mounted
	// here apply to all the service endpoints.
	var handler http.Handler = mux
	{
		handler = httpmdlwr.RequestID()(handler)
		handler = customZeroLog(logger)(handler)
	}

	// Start HTTP server using default configuration, change the code to
	// configure the server as required by your service.
	srv := &http.Server{Addr: u.Host, Handler: handler}
	for _, m := range frontServer.Mounts {
		logger.Debug().Msg(fmt.Sprintf("HTTP %q mounted on %s %s", m.Method, m.Verb, m.Pattern))
	}

	(*wg).Add(1)
	go func() {
		defer (*wg).Done()

		// Start HTTP server in a separate goroutine.
		go func() {
			logger.Info().Msg(fmt.Sprintf("HTTP server listening on %q", u.Host))
			errc <- srv.ListenAndServe()
		}()

		<-ctx.Done()
		logger.Info().Msg(fmt.Sprintf("shutting down HTTP server at %q", u.Host))

		// Shutdown gracefully with a 30s timeout.
		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()

		_ = srv.Shutdown(ctx)
	}()
}

// errorHandler returns a function that writes and logs the given error.
// The function also writes and logs the error unique ID so that it's possible
// to correlate.
func errorHandler(logger *zerolog.Logger) func(context.Context, http.ResponseWriter, error) {
	return func(ctx context.Context, w http.ResponseWriter, err error) {
		fmt.Println("TEST HERE!****************************************")
		id := ctx.Value(middleware.RequestIDKey).(string)
		_, _ = w.Write([]byte("[" + id + "] encoding: " + err.Error()))
		//logger.Printf("[%s] ERROR: %s", id, err.Error())
		logger.Error().Stack().Err(err).Msg("")
	}
}
