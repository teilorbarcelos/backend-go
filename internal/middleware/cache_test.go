package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

func TestCacheControl(t *testing.T) {
	t.Run("GET JSON returns ETag and Cache-Control", func(t *testing.T) {
		r := gin.New()
		r.Use(CacheControl())
		r.GET("/test", func(c *gin.Context) {
			c.JSON(http.StatusOK, gin.H{"data": "test"})
		})

		w := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodGet, "/test", nil)
		r.ServeHTTP(w, req)

		assert.NotEmpty(t, w.Header().Get("ETag"))
		assert.Equal(t, "no-cache, must-revalidate", w.Header().Get("Cache-Control"))
	})

	t.Run("GET with matching ETag returns 304", func(t *testing.T) {
		r := gin.New()
		r.Use(CacheControl())
		r.GET("/test", func(c *gin.Context) {
			c.JSON(http.StatusOK, gin.H{"data": "test"})
		})

		w := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodGet, "/test", nil)
		r.ServeHTTP(w, req)

		etag := w.Header().Get("ETag")

		r2 := gin.New()
		r2.Use(CacheControl())
		r2.GET("/test", func(c *gin.Context) {
			c.JSON(http.StatusOK, gin.H{"data": "test"})
		})

		w2 := httptest.NewRecorder()
		req2 := httptest.NewRequest(http.MethodGet, "/test", nil)
		req2.Header.Set("If-None-Match", etag)
		r2.ServeHTTP(w2, req2)

		assert.Equal(t, http.StatusNotModified, w2.Code)
	})

	t.Run("POST is not cached", func(t *testing.T) {
		r := gin.New()
		r.Use(CacheControl())
		r.POST("/test", func(c *gin.Context) {
			c.JSON(http.StatusOK, gin.H{"data": "test"})
		})

		w := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodPost, "/test", nil)
		r.ServeHTTP(w, req)

		assert.Empty(t, w.Header().Get("ETag"))
	})

	t.Run("Non-200 status passes through", func(t *testing.T) {
		r := gin.New()
		r.Use(CacheControl())
		r.GET("/test", func(c *gin.Context) {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "fail"})
		})

		w := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodGet, "/test", nil)
		r.ServeHTTP(w, req)

		assert.Empty(t, w.Header().Get("ETag"))
		assert.Equal(t, http.StatusInternalServerError, w.Code)
		assert.Contains(t, w.Body.String(), "fail")
	})

	t.Run("Non-JSON response passes through without ETag", func(t *testing.T) {
		r := gin.New()
		r.Use(CacheControl())
		r.GET("/test", func(c *gin.Context) {
			c.Data(http.StatusOK, "text/plain", []byte("hello world"))
		})

		w := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodGet, "/test", nil)
		r.ServeHTTP(w, req)

		assert.Empty(t, w.Header().Get("ETag"))
		assert.Equal(t, http.StatusOK, w.Code)
		assert.Equal(t, "hello world", w.Body.String())
	})
}
