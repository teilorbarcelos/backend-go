package pdf

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

// PdfRequestDTO defines the payload for PDF generation.
type PdfRequestDTO struct {
	Template string                 `json:"template"`
	Data     map[string]interface{} `json:"data"`
	Options  Options                `json:"options"`
}

// Options defines the PDF generation options.
type Options struct {
	Landscape bool   `json:"landscape"`
	Format    string `json:"format"`
}

// PdfProvider defines the interface for PDF generation.
type PdfProvider interface {
	GeneratePdf(request PdfRequestDTO) (io.ReadCloser, error)
}

// RemotePdfProvider implements PdfProvider calling a remote service.
type RemotePdfProvider struct {
	ServiceURL string
	HTTPClient *http.Client
}

// NewRemotePdfProvider creates a new RemotePdfProvider.
func NewRemotePdfProvider(serviceURL string) *RemotePdfProvider {
	return &RemotePdfProvider{
		ServiceURL: serviceURL,
		HTTPClient: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

// GeneratePdf calls the remote service and returns the response body as a stream.
func (p *RemotePdfProvider) GeneratePdf(request PdfRequestDTO) (io.ReadCloser, error) {
	jsonData, err := json.Marshal(request)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	resp, err := p.HTTPClient.Post(p.ServiceURL+"/v1/pdf/generate", "application/json", bytes.NewBuffer(jsonData))
	if err != nil {
		return nil, fmt.Errorf("failed to call pdf service: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		resp.Body.Close()
		return nil, fmt.Errorf("pdf service returned status: %s", resp.Status)
	}

	return resp.Body, nil
}
