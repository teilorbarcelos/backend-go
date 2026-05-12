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

//go:embed templates/*
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

func automateRegistration(data TemplateData) {
	// 1. Factory Code
	factoryPath := filepath.Join("pkg", "storage", "factory.go")
	factoryContent, err := os.ReadFile(factoryPath)
	if err == nil {
		content := string(factoryContent)
		caseStr := fmt.Sprintf("case \"%s\":\n\t\treturn New%sProvider(bucket), nil", data.LowerName, data.Name)
		if !strings.Contains(content, caseStr) {
			newCase := caseStr + "\n\tdefault:"
			content = strings.Replace(content, "default:", newCase, 1)
			os.WriteFile(factoryPath, []byte(content), 0644)
			fmt.Println("Registrado Provider na Factory em pkg/storage/factory.go")
		}
	}

	// 2. Factory Test
	testPath := filepath.Join("pkg", "storage", "factory_test.go")
	if _, err := os.Stat(testPath); os.IsNotExist(err) {
		writeTemplate(testPath, "factory_test_base.tpl", data)
	}

	testContent, err := os.ReadFile(testPath)
	if err == nil {
		content := string(testContent)
		testCase := fmt.Sprintf("t.Run(\"Success %s\", func(t *testing.T) {\n\t\tprovider, err := NewStorageProvider(\"%s\", \"bucket\")\n\t\tassert.NoError(t, err)\n\t\tassert.NotNil(t, provider)\n\t})", data.Name, data.LowerName)
		if !strings.Contains(content, "Success "+data.Name) {
			newTest := testCase + "\n\n\tt.Run(\"Unsupported Driver\""
			content = strings.Replace(content, "t.Run(\"Unsupported Driver\"", newTest, 1)
			os.WriteFile(testPath, []byte(content), 0644)
			fmt.Println("Registrado Teste na Factory em pkg/storage/factory_test.go")
		}
	}
}

func main() {
	if len(os.Args) < 2 {
		log.Fatal("Uso: go run tools/generator/storage/main.go <ProviderName>")
	}

	name := os.Args[1]
	// Garantir PascalCase
	if len(name) > 0 {
		name = strings.ToUpper(name[:1]) + name[1:]
	}
	lowerName := strings.ToLower(name)

	dir := filepath.Join("pkg", "storage")
	if err := os.MkdirAll(dir, 0755); err != nil {
		log.Fatalf("Erro ao criar diretório %s: %v", dir, err)
	}

	data := TemplateData{
		Name:      name,
		LowerName: lowerName,
	}

	// Arquivo do provider
	writeTemplate(filepath.Join(dir, lowerName+".go"), lowerName+".tpl", data)
	// Arquivo de teste individual do provider
	writeTemplate(filepath.Join(dir, lowerName+"_test.go"), "storage_test.tpl", data)

	// Automação (Factory e Factory Test)
	automateRegistration(data)

	fmt.Printf("\nDriver de storage '%s' instalado e registrado com sucesso!\n", name)
}
