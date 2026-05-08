package main

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"
)

var storageTemplates = map[string]string{
	"s3": `package storage

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
`,
	"gcs": `package storage

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
`,
	"azure": `package storage

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
`,
}

func main() {
	if len(os.Args) < 2 {
		log.Fatal("Uso: go run tools/generator/install-storage.go <ProviderName>")
	}

	name := strings.ToLower(os.Args[1])

	tmpl, exists := storageTemplates[name]
	if !exists {
		log.Fatalf("Provider de storage '%s' não suportado. Opções: s3, gcs, azure", name)
	}

	dir := filepath.Join("pkg", "storage")
	if err := os.MkdirAll(dir, 0755); err != nil {
		log.Fatalf("Erro ao criar diretório %s: %v", dir, err)
	}

	filePath := filepath.Join(dir, name+".go")
	if err := os.WriteFile(filePath, []byte(tmpl), 0644); err != nil {
		log.Fatalf("Erro ao escrever arquivo %s: %v", filePath, err)
	}

	fmt.Printf("\nDriver de storage '%s' instalado com sucesso em '%s'.\n", name, filePath)
}
