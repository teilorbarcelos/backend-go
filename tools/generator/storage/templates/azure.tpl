package storage

import "log"

type AzureDriver struct{}

func NewAzureDriver() *AzureDriver {
	log.Println("AzureDriver inicializado")
	return &AzureDriver{}
}

func (d *AzureDriver) Upload(filename string, data []byte) error {
	log.Printf("Upload para Azure Blob: %s\n", filename)
	return nil
}
