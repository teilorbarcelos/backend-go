package storage

import "log"

type GCSDriver struct{}

func NewGCSDriver() *GCSDriver {
	log.Println("GCSDriver inicializado")
	return &GCSDriver{}
}

func (d *GCSDriver) Upload(filename string, data []byte) error {
	log.Printf("Upload para GCS: %s\n", filename)
	return nil
}
