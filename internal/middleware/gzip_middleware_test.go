package middleware

import (
	"bytes"
	"compress/gzip"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/require"
)

func testHandler(contentType, body string) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", contentType)
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(body))
	})
}

func TestGzip_CompressJSON(t *testing.T) {
	handler := GzipMiddleware(testHandler("application/json", `{"result":"ok"}`))

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set("Accept-Encoding", "gzip")
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	res := rec.Result()
	defer res.Body.Close()

	require.Equal(t, "gzip", res.Header.Get("Content-Encoding"))

	gr, err := gzip.NewReader(res.Body)
	require.NoError(t, err)
	defer gr.Close()

	b, err := io.ReadAll(gr)
	require.NoError(t, err)
	require.Equal(t, `{"result":"ok"}`, string(b))
}

func TestGzip_CompressHTML(t *testing.T) {
	handler := GzipMiddleware(testHandler("text/html", "<html></html>"))

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set("Accept-Encoding", "gzip")
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	res := rec.Result()
	defer res.Body.Close()

	require.Equal(t, "gzip", res.Header.Get("Content-Encoding"))

	gr, err := gzip.NewReader(res.Body)
	require.NoError(t, err)
	defer gr.Close()

	b, err := io.ReadAll(gr)
	require.NoError(t, err)
	require.Equal(t, "<html></html>", string(b))
}

func TestGzip_NoCompressPlainText(t *testing.T) {
	handler := GzipMiddleware(testHandler("text/plain", "hello"))

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set("Accept-Encoding", "gzip")
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	res := rec.Result()
	defer res.Body.Close()

	b, err := io.ReadAll(res.Body)
	require.NoError(t, err)
	require.Equal(t, "hello", string(b))
}

func TestGzip_NoCompressWithoutAcceptEncoding(t *testing.T) {
	handler := GzipMiddleware(testHandler("application/json", `{"ok":true}`))

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	res := rec.Result()
	defer res.Body.Close()

	require.Empty(t, res.Header.Get("Content-Encoding"))

	b, err := io.ReadAll(res.Body)
	require.NoError(t, err)
	require.Equal(t, `{"ok":true}`, string(b))
}

func TestGzip_DecompressRequest(t *testing.T) {
	inner := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		b, _ := io.ReadAll(r.Body)
		w.Header().Set("Content-Type", "text/plain")
		w.Write(b)
	})
	handler := GzipMiddleware(inner)

	var buf bytes.Buffer
	gz := gzip.NewWriter(&buf)
	gz.Write([]byte("compressed body"))
	gz.Close()

	req := httptest.NewRequest(http.MethodPost, "/", &buf)
	req.Header.Set("Content-Encoding", "gzip")
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	res := rec.Result()
	defer res.Body.Close()

	b, err := io.ReadAll(res.Body)
	require.NoError(t, err)
	require.Equal(t, "compressed body", string(b))
}

func TestGzip_NoCompressImage(t *testing.T) {
	handler := GzipMiddleware(testHandler("image/png", "fake-png-data"))

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set("Accept-Encoding", "gzip")
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	res := rec.Result()
	defer res.Body.Close()

	b, err := io.ReadAll(res.Body)
	require.NoError(t, err)
	require.Equal(t, "fake-png-data", string(b))
}
