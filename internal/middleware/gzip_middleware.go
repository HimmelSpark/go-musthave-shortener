package middleware

import (
	"bytes"
	"compress/gzip"
	"fmt"
	"io"
	"net/http"
	"strings"
)

var compressibleTypes = map[string]bool{
	"application/json": true,
	"text/html":        true,
}

func shouldCompress(contentType string) bool {
	ct := strings.SplitN(contentType, ";", 2)[0]
	return compressibleTypes[strings.TrimSpace(ct)]
}

type gzipReadCloser struct {
	io.Reader
	body io.Closer
	gzip io.Closer
}

func (g *gzipReadCloser) Close() error {
	var errs []error
	if g.gzip != nil {
		if err := g.gzip.Close(); err != nil {
			errs = append(errs, err)
		}
	}
	if g.body != nil {
		if err := g.body.Close(); err != nil {
			errs = append(errs, err)
		}
	}
	if len(errs) == 0 {
		return nil
	}
	return fmt.Errorf("close gzip body: %w", errorsJoin(errs...))
}

func errorsJoin(errs ...error) error {
	var result error
	for _, err := range errs {
		if err == nil {
			continue
		}
		if result == nil {
			result = err
			continue
		}
		result = fmt.Errorf("%v; %w", result, err)
	}
	return result
}

type bufferedResponseWriter struct {
	http.ResponseWriter
	buf         bytes.Buffer
	status      int
	wroteHeader bool
}

func (b *bufferedResponseWriter) WriteHeader(statusCode int) {
	if b.wroteHeader {
		return
	}
	b.status = statusCode
	b.wroteHeader = true
}

func (b *bufferedResponseWriter) Write(data []byte) (int, error) {
	if !b.wroteHeader {
		b.WriteHeader(http.StatusOK)
	}
	return b.buf.Write(data)
}

func GzipMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("Content-Encoding") == "gzip" {
			gr, err := gzip.NewReader(r.Body)
			if err != nil {
				http.Error(w, "failed to decompress request", http.StatusBadRequest)
				return
			}
			r.Body = &gzipReadCloser{Reader: gr, body: r.Body, gzip: gr}
		}

		acceptsGzip := false
		for _, encoding := range strings.Split(r.Header.Get("Accept-Encoding"), ",") {
			if strings.EqualFold(strings.TrimSpace(encoding), "gzip") {
				acceptsGzip = true
				break
			}
		}

		if !acceptsGzip {
			next.ServeHTTP(w, r)
			return
		}

		bw := &bufferedResponseWriter{ResponseWriter: w, status: http.StatusOK}
		next.ServeHTTP(bw, r)

		body := bw.buf.Bytes()
		ct := bw.Header().Get("Content-Type")
		if ct == "" && len(body) > 0 {
			ct = http.DetectContentType(body)
			bw.Header().Set("Content-Type", ct)
		}

		if shouldCompress(ct) && len(body) > 0 {
			var compressed bytes.Buffer
			gz, err := gzip.NewWriterLevel(&compressed, gzip.BestSpeed)
			if err == nil {
				_, err = gz.Write(body)
				if closeErr := gz.Close(); err == nil {
					err = closeErr
				}
			}
			if err == nil {
				w.Header().Set("Content-Encoding", "gzip")
				w.Header().Del("Content-Length")
				w.WriteHeader(bw.status)
				_, _ = w.Write(compressed.Bytes())
				return
			}
		}

		w.Header().Del("Content-Encoding")
		w.WriteHeader(bw.status)
		_, _ = w.Write(body)
	})
}
