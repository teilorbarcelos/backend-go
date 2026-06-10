package middleware

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
)

type cacheResponseWriter struct {
	gin.ResponseWriter
	body   *bytes.Buffer
	status int
}

func (w *cacheResponseWriter) WriteHeader(status int) {
	w.status = status
}

func (w *cacheResponseWriter) Write(data []byte) (int, error) {
	return w.body.Write(data)
}

func CacheControl() gin.HandlerFunc {
	return func(c *gin.Context) {
		if c.Request.Method != http.MethodGet {
			c.Next()
			return
		}

		w := &cacheResponseWriter{ResponseWriter: c.Writer, body: &bytes.Buffer{}}
		c.Writer = w

		c.Next()

		if w.body.Len() == 0 || w.status == 0 {
			w.ResponseWriter.WriteHeader(w.status)
			w.body.WriteTo(w.ResponseWriter)
			return
		}

		contentType := c.Writer.Header().Get("Content-Type")

		if w.status != http.StatusOK || !strings.HasPrefix(contentType, "application/json") {
			w.ResponseWriter.WriteHeader(w.status)
			w.body.WriteTo(w.ResponseWriter)
			return
		}

		hash := sha256.Sum256(w.body.Bytes())
		etag := `"` + hex.EncodeToString(hash[:]) + `"`

		c.Header("ETag", etag)
		c.Header("Cache-Control", "no-cache, must-revalidate")

		if match := c.GetHeader("If-None-Match"); match == etag {
			w.ResponseWriter.WriteHeader(http.StatusNotModified)
			return
		}

		w.ResponseWriter.WriteHeader(w.status)
		w.body.WriteTo(w.ResponseWriter)
	}
}
