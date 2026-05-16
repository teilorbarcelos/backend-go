package debug

import (
	"bytes"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"backend-go/internal/infra/pdf"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// MockPdfProvider is a mock implementation of pdf.PdfProvider
type MockPdfProvider struct {
	mock.Mock
}

func (m *MockPdfProvider) GeneratePdf(request pdf.PdfRequestDTO) (io.ReadCloser, error) {
	args := m.Called(request)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(io.ReadCloser), args.Error(1)
}

func TestRegisterRoutes(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	v1 := r.Group("/v1")
	RegisterRoutes(v1)

	routes := r.Routes()
	assert.Len(t, routes, 2)

	foundPost := false
	foundGet := false
	for _, route := range routes {
		if route.Path == "/v1/debug/pdf" && route.Method == "POST" {
			foundPost = true
		}
		if route.Path == "/v1/debug/pdf" && route.Method == "GET" {
			foundGet = true
		}
	}
	assert.True(t, foundPost)
	assert.True(t, foundGet)
}

func TestDebugHandler_TestPdf(t *testing.T) {
	gin.SetMode(gin.TestMode)

	t.Run("success", func(t *testing.T) {
		mockProvider := new(MockPdfProvider)
		handler := &DebugHandler{pdfProvider: mockProvider}

		requestDTO := pdf.PdfRequestDTO{Template: "test"}
		pdfContent := "%PDF-test"
		mockProvider.On("GeneratePdf", requestDTO).Return(io.NopCloser(strings.NewReader(pdfContent)), nil)

		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		
		body, _ := io.ReadAll(strings.NewReader(`{"template":"test"}`))
		c.Request = httptest.NewRequest("POST", "/v1/debug/pdf", bytes.NewBuffer(body))
		c.Request.Header.Set("Content-Type", "application/json")

		handler.TestPdf(c)

		assert.Equal(t, http.StatusOK, w.Code)
		assert.Equal(t, "application/pdf", w.Header().Get("Content-Type"))
		assert.Equal(t, "attachment; filename=\"test.pdf\"", w.Header().Get("Content-Disposition"))
		assert.Equal(t, pdfContent, w.Body.String())
	})

	t.Run("bind_error", func(t *testing.T) {
		handler := &DebugHandler{}

		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = httptest.NewRequest("POST", "/v1/debug/pdf", strings.NewReader("invalid json"))
		c.Request.Header.Set("Content-Type", "application/json")

		handler.TestPdf(c)

		assert.Equal(t, http.StatusBadRequest, w.Code)
		assert.Contains(t, w.Body.String(), "error")
	})

	t.Run("provider_error", func(t *testing.T) {
		mockProvider := new(MockPdfProvider)
		handler := &DebugHandler{pdfProvider: mockProvider}

		mockProvider.On("GeneratePdf", mock.Anything).Return(nil, errors.New("provider error"))

		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = httptest.NewRequest("POST", "/v1/debug/pdf", strings.NewReader(`{"template":"test"}`))
		c.Request.Header.Set("Content-Type", "application/json")

		handler.TestPdf(c)

		assert.Equal(t, http.StatusInternalServerError, w.Code)
		assert.Contains(t, w.Body.String(), "provider error")
	})

	t.Run("io_copy_error", func(t *testing.T) {
		mockProvider := new(MockPdfProvider)
		handler := &DebugHandler{pdfProvider: mockProvider}

		mockProvider.On("GeneratePdf", mock.Anything).Return(io.NopCloser(strings.NewReader("%PDF-test")), nil)

		w := &errResponseWriter{}
		c, _ := gin.CreateTestContext(w)
		c.Request = httptest.NewRequest("POST", "/v1/debug/pdf", strings.NewReader(`{"template":"test"}`))
		c.Request.Header.Set("Content-Type", "application/json")

		handler.TestPdf(c)
		// We can't check status code here as it might have been set to 200 before copy
	})
}

func TestDebugHandler_TestPdfGet(t *testing.T) {
	gin.SetMode(gin.TestMode)

	t.Run("success", func(t *testing.T) {
		mockProvider := new(MockPdfProvider)
		handler := &DebugHandler{pdfProvider: mockProvider}

		pdfContent := "%PDF-get-test"
		mockProvider.On("GeneratePdf", mock.Anything).Return(io.NopCloser(strings.NewReader(pdfContent)), nil)

		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = httptest.NewRequest("GET", "/v1/debug/pdf", nil)

		handler.TestPdfGet(c)

		assert.Equal(t, http.StatusOK, w.Code)
		assert.Equal(t, "application/pdf", w.Header().Get("Content-Type"))
		assert.Equal(t, "inline; filename=\"test.pdf\"", w.Header().Get("Content-Disposition"))
		assert.Equal(t, pdfContent, w.Body.String())
	})

	t.Run("provider_error", func(t *testing.T) {
		mockProvider := new(MockPdfProvider)
		handler := &DebugHandler{pdfProvider: mockProvider}

		mockProvider.On("GeneratePdf", mock.Anything).Return(nil, errors.New("provider error"))

		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = httptest.NewRequest("GET", "/v1/debug/pdf", nil)

		handler.TestPdfGet(c)

		assert.Equal(t, http.StatusInternalServerError, w.Code)
		assert.Contains(t, w.Body.String(), "provider error")
	})

	t.Run("io_copy_error", func(t *testing.T) {
		mockProvider := new(MockPdfProvider)
		handler := &DebugHandler{pdfProvider: mockProvider}

		mockProvider.On("GeneratePdf", mock.Anything).Return(io.NopCloser(strings.NewReader("%PDF-test")), nil)

		w := &errResponseWriter{}
		c, _ := gin.CreateTestContext(w)
		c.Request = httptest.NewRequest("GET", "/v1/debug/pdf", nil)

		handler.TestPdfGet(c)
	})
}

type errResponseWriter struct {
	header http.Header
}

func (w *errResponseWriter) Header() http.Header {
	if w.header == nil {
		w.header = make(http.Header)
	}
	return w.header
}

func (w *errResponseWriter) Write(b []byte) (int, error) {
	return 0, errors.New("write error")
}

func (w *errResponseWriter) WriteHeader(statusCode int) {}
