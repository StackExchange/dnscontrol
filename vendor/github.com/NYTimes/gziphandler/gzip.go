package gziphandler

import (
	"compress/gzip"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"sync"
)

const (
	vary            = "Vary"
	acceptEncoding  = "Accept-Encoding"
	contentEncoding = "Content-Encoding"
)

type codings map[string]float64

// The default qvalue to assign to an encoding if no explicit qvalue is set.
// This is actually kind of ambiguous in RFC 2616, so hopefully it's correct.
// The examples seem to indicate that it is.
const DEFAULT_QVALUE = 1.0

// gzipWriterPools stores a sync.Pool for each compression level for re-uze of gzip.Writers.
// Use poolIndex to covert a compression level to an index into gzipWriterPools.
var gzipWriterPools [gzip.BestCompression - gzip.BestSpeed + 2]*sync.Pool

func init() {
	for i := gzip.BestSpeed; i <= gzip.BestCompression; i++ {
		addLevelPool(i)
	}
	addLevelPool(gzip.DefaultCompression)
}

// poolIndex maps a compression level to its index into gzipWriterPools. It assumes that
// level is a valid gzip compression level.
func poolIndex(level int) int {
	// gzip.DefaultCompression == -1, so we need to treat it special.
	if level == gzip.DefaultCompression {
		return gzip.BestCompression - gzip.BestSpeed + 1
	}
	return level - gzip.BestSpeed
}

func addLevelPool(level int) {
	gzipWriterPools[poolIndex(level)] = &sync.Pool{
		New: func() interface{} {
			// NewWriterLevel only returns error on a bad level, we are guaranteeing
			// that this will be a valid level so it is okay to ignore the returned
			// error.
			w, _ := gzip.NewWriterLevel(nil, level)
			return w
		},
	}
}

// GzipResponseWriter provides an http.ResponseWriter interface, which gzips
// bytes before writing them to the underlying response. This doesn't set the
// Content-Encoding header, nor close the writers, so don't forget to do that.
type GzipResponseWriter struct {
	gw *gzip.Writer
	http.ResponseWriter
}

// Write appends data to the gzip writer.
func (w GzipResponseWriter) Write(b []byte) (int, error) {
	if _, ok := w.Header()["Content-Type"]; !ok {
		// If content type is not set, infer it from the uncompressed body.
		w.Header().Set("Content-Type", http.DetectContentType(b))
	}
	return w.gw.Write(b)
}

// Flush flushes the underlying *gzip.Writer and then the underlying
// http.ResponseWriter if it is an http.Flusher. This makes GzipResponseWriter
// an http.Flusher.
func (w GzipResponseWriter) Flush() {
	w.gw.Flush()
	if fw, ok := w.ResponseWriter.(http.Flusher); ok {
		fw.Flush()
	}
}

// MustNewGzipLevelHandler behaves just like NewGzipLevelHandler except that in an error case
// it panics rather than returning an error.
func MustNewGzipLevelHandler(level int) func(http.Handler) http.Handler {
	wrap, err := NewGzipLevelHandler(level)
	if err != nil {
		panic(err)
	}
	return wrap
}

// NewGzipLevelHandler returns a wrapper function (often known as middleware)
// which can be used to wrap an HTTP handler to transparently gzip the response
// body if the client supports it (via the Accept-Encoding header). Responses will
// be encoded at the given gzip compression level. An error will be returned only
// if an invalid gzip compression level is given, so if one can ensure the level
// is valid, the returned error can be safely ignored.
func NewGzipLevelHandler(level int) (func(http.Handler) http.Handler, error) {
	if level != gzip.DefaultCompression && (level < gzip.BestSpeed || level > gzip.BestCompression) {
		return nil, fmt.Errorf("invalid compression level requested: %d", level)
	}
	return func(h http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Header().Add(vary, acceptEncoding)

			if acceptsGzip(r) {
				// Bytes written during ServeHTTP are redirected to this gzip writer
				// before being written to the underlying response.
				gzw := gzipWriterPools[poolIndex(level)].Get().(*gzip.Writer)
				defer gzipWriterPools[poolIndex(level)].Put(gzw)
				gzw.Reset(w)
				defer gzw.Close()

				w.Header().Set(contentEncoding, "gzip")
				h.ServeHTTP(GzipResponseWriter{gzw, w}, r)
			} else {
				h.ServeHTTP(w, r)
			}
		})
	}, nil
}

// GzipHandler wraps an HTTP handler, to transparently gzip the response body if
// the client supports it (via the Accept-Encoding header). This will compress at
// the default compression level.
func GzipHandler(h http.Handler) http.Handler {
	wrapper, _ := NewGzipLevelHandler(gzip.DefaultCompression)
	return wrapper(h)
}

// acceptsGzip returns true if the given HTTP request indicates that it will
// accept a gzippped response.
func acceptsGzip(r *http.Request) bool {
	acceptedEncodings, _ := parseEncodings(r.Header.Get(acceptEncoding))
	return acceptedEncodings["gzip"] > 0.0
}

// parseEncodings attempts to parse a list of codings, per RFC 2616, as might
// appear in an Accept-Encoding header. It returns a map of content-codings to
// quality values, and an error containing the errors encounted. It's probably
// safe to ignore those, because silently ignoring errors is how the internet
// works.
//
// See: http://tools.ietf.org/html/rfc2616#section-14.3
func parseEncodings(s string) (codings, error) {
	c := make(codings)
	e := make([]string, 0)

	for _, ss := range strings.Split(s, ",") {
		coding, qvalue, err := parseCoding(ss)

		if err != nil {
			e = append(e, err.Error())

		} else {
			c[coding] = qvalue
		}
	}

	// TODO (adammck): Use a proper multi-error struct, so the individual errors
	//                 can be extracted if anyone cares.
	if len(e) > 0 {
		return c, fmt.Errorf("errors while parsing encodings: %s", strings.Join(e, ", "))
	}

	return c, nil
}

// parseCoding parses a single conding (content-coding with an optional qvalue),
// as might appear in an Accept-Encoding header. It attempts to forgive minor
// formatting errors.
func parseCoding(s string) (coding string, qvalue float64, err error) {
	for n, part := range strings.Split(s, ";") {
		part = strings.TrimSpace(part)
		qvalue = DEFAULT_QVALUE

		if n == 0 {
			coding = strings.ToLower(part)

		} else if strings.HasPrefix(part, "q=") {
			qvalue, err = strconv.ParseFloat(strings.TrimPrefix(part, "q="), 64)

			if qvalue < 0.0 {
				qvalue = 0.0

			} else if qvalue > 1.0 {
				qvalue = 1.0
			}
		}
	}

	if coding == "" {
		err = fmt.Errorf("empty content-coding")
	}

	return
}
