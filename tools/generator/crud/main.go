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

type TemplateData struct {
	Name      string
	LowerName string
}

func writeTemplate(outputPath, templateName string, data TemplateData) {
	tmplContent, err := templatesFS.ReadFile(filepath.Join("templates", templateName))
	if err != nil {
		log.Fatalf("Erro ao ler template %s: %v", templateName, err)
	}

	t, err := template.New(templateName).Parse(string(tmplContent))
	if err != nil {
		log.Fatalf("Erro ao parsear template %s: %v", templateName, err)
	}

	var buf bytes.Buffer
	if err := t.Execute(&buf, data); err != nil {
		log.Fatalf("Erro ao executar template %s: %v", templateName, err)
	}

	if err := os.WriteFile(outputPath, buf.Bytes(), 0644); err != nil {
		log.Fatalf("Erro ao escrever arquivo %s: %v", outputPath, err)
	}
	fmt.Printf("Criado: %s\n", outputPath)
}

func main() {
	if len(os.Args) < 2 {
		log.Fatal("Uso: go run tools/generator/crud/main.go <ModuleName>")
	}

	name := os.Args[1]
	lowerName := strings.ToLower(name)

	dir := filepath.Join("internal", "app", lowerName)
	if err := os.MkdirAll(dir, 0755); err != nil {
		log.Fatalf("Erro ao criar diretório %s: %v", dir, err)
	}

	data := TemplateData{
		Name:      name,
		LowerName: lowerName,
	}

	writeTemplate(filepath.Join(dir, "repository.go"), "repository.tpl", data)
	writeTemplate(filepath.Join(dir, "service.go"), "service.tpl", data)
	writeTemplate(filepath.Join(dir, "handler.go"), "handler.tpl", data)
	writeTemplate(filepath.Join(dir, "repository_test.go"), "repository_test.tpl", data)
	writeTemplate(filepath.Join(dir, "service_test.go"), "service_test.tpl", data)
	writeTemplate(filepath.Join(dir, "handler_test.go"), "handler_test.tpl", data)

	fmt.Printf("\nMódulo '%s' gerado com sucesso em '%s'.\n", name, dir)
}
