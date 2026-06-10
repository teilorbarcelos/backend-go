package middleware

import (
	"compress/gzip"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

func TestGzip(t *testing.T) {
	t.Run("compresses response when client accepts gzip", func(t *testing.T) {
		r := gin.New()
		r.Use(Gzip())
		r.GET("/test", func(c *gin.Context) {
			c.String(http.StatusOK, "hello world")
		})

		w := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodGet, "/test", nil)
		req.Header.Set("Accept-Encoding", "gzip")
		r.ServeHTTP(w, req)

		assert.Equal(t, "gzip", w.Header().Get("Content-Encoding"))

		reader, err := gzip.NewReader(w.Body)
		assert.NoError(t, err)
		defer reader.Close()

		data, err := io.ReadAll(reader)
		assert.NoError(t, err)
		assert.Equal(t, "hello world", string(data))
	})

	t.Run("gzip writer error falls through", func(t *testing.T) {
		orig := gzipNewWriter
		gzipNewWriter = func(w io.Writer) (*gzip.Writer, error) {
			return nil, errors.New("gzip error")
		}
		defer func() { gzipNewWriter = orig }()

		r := gin.New()
		r.Use(Gzip())
		r.GET("/test", func(c *gin.Context) {
			c.String(http.StatusOK, "hello world")
		})

		w := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodGet, "/test", nil)
		req.Header.Set("Accept-Encoding", "gzip")
		r.ServeHTTP(w, req)

		assert.Empty(t, w.Header().Get("Content-Encoding"))
		assert.Equal(t, "hello world", w.Body.String())
	})

	t.Run("does not compress when client does not accept gzip", func(t *testing.T) {
		r := gin.New()
		r.Use(Gzip())
		r.GET("/test", func(c *gin.Context) {
			c.String(http.StatusOK, "hello world")
		})

		w := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodGet, "/test", nil)
		r.ServeHTTP(w, req)

		assert.Empty(t, w.Header().Get("Content-Encoding"))
		assert.Equal(t, "hello world", w.Body.String())
	})
}
