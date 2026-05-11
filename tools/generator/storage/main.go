package main

import (
	"bytes"
	"embed"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"
	"text/template"
)

var templatesFS embed.FS

func main() {
	if len(os.Args) < 2 {
		log.Fatal("Uso: go run tools/generator/storage/main.go <ProviderName>")
	}

	name := strings.ToLower(os.Args[1])
	templateFile := name + ".tpl"

	tmplContent, err := templatesFS.ReadFile(filepath.Join("templates", templateFile))
	if err != nil {
		log.Fatalf("Provider de storage '%s' não suportado. Opções: s3, gcs, azure", name)
	}

	dir := filepath.Join("pkg", "storage")
	if err := os.MkdirAll(dir, 0755); err != nil {
		log.Fatalf("Erro ao criar diretório %s: %v", dir, err)
	}

	filePath := filepath.Join(dir, name+".go")

	t, err := template.New(templateFile).Parse(string(tmplContent))
	if err != nil {
		log.Fatalf("Erro ao parsear template %s: %v", templateFile, err)
	}

	var buf bytes.Buffer
	if err := t.Execute(&buf, nil); err != nil {
		log.Fatalf("Erro ao executar template %s: %v", templateFile, err)
	}

	if err := os.WriteFile(filePath, buf.Bytes(), 0644); err != nil {
		log.Fatalf("Erro ao escrever arquivo %s: %v", filePath, err)
	}

	fmt.Printf("\nDriver de storage '%s' instalado com sucesso em '%s'.\n", name, filePath)
}
