package storage

import "log"

type S3Driver struct{}

func NewS3Driver() *S3Driver {
	log.Println("S3Driver inicializado")
	return &S3Driver{}
}

func (d *S3Driver) Upload(filename string, data []byte) error {
	log.Printf("Upload para S3: %s\n", filename)
	return nil
}
