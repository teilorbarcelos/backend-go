# 🚀 Go Modular Backend Boilerplate

[![Go Coverage](https://img.shields.io/badge/Coverage-100%25-brightgreen.svg)](#-testes-e-cobertura)
[![Go Report Card](https://goreportcard.com/badge/github.com/teilorbarcelos/backend-go)](https://goreportcard.com/report/github.com/teilorbarcelos/backend-go)
[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](LICENSE)
[![CI](https://github.com/teilorbarcelos/backend-go/actions/workflows/ci.yml/badge.svg)](https://github.com/teilorbarcelos/backend-go/actions/workflows/ci.yml)

Este é um boilerplate de nível **Enterprise** construído em Go, seguindo os princípios de **Clean Architecture** e **Modular Monolith**. Projetado para alta escalabilidade, performance extrema e uma experiência de desenvolvimento (DX) superior.

---

## 💎 Diferenciais Estratégicos

Este projeto não é apenas uma estrutura de pastas, mas um ecossistema completo para aplicações de missão crítica:

*   **🛡️ Segurança Industrial:** RBAC (Role-Based Access Control), JWT com invalidação via Redis, Rate Limiting e Auditoria Automática (Audit Logs).
*   **🏗️ Geradores de Código (Scaffolding):** Crie módulos CRUD completos ou novos Storage Drivers em segundos com comandos CLI.
*   **🧪 Testes de Alta Fidelidade:** 100% de cobertura de código garantida por **Testcontainers** (Postgres e Redis reais nos testes).
*   **📊 Observabilidade Nativa:** Middleware de métricas Prometheus e Logging estruturado com Zap.
*   **📂 Gestão de Media Modular:** Factory pattern para múltiplos storages (S3, Local, etc.) com geração automática de infraestrutura.

---

## 🛠️ Stack Tecnológica

*   **Linguagem:** Go 1.21+
*   **Web Framework:** Gin Gonic
*   **ORM:** GORM (PostgreSQL)
*   **Cache/Session:** Redis
*   **Mensageria:** RabbitMQ
*   **Documentação:** Swagger / OpenAPI
*   **Infra Dev:** Docker & Docker Compose
*   **Live Reload:** Air

---

## 🏗️ Arquitetura

O projeto utiliza uma abordagem de **Modular Monolith**, permitindo que o sistema cresça de forma organizada e seja facilmente decomposto em microserviços se necessário.

```text
.
├── cmd/api/            # Ponto de entrada da aplicação
├── internal/
│   ├── app/            # Módulos de Domínio (User, Role, Media, etc.)
│   ├── core/           # Lógica compartilhada (Models, Audit, Middleware)
│   └── infra/          # Implementações de infraestrutura (Session, etc.)
├── pkg/                # Pacotes utilitários reutilizáveis (Database, Logger, Security)
├── tools/              # Ferramentas auxiliares e Geradores
└── docker-compose.yml  # Orquestração de serviços
```

---

## 🚀 Iniciando o Ambiente de Desenvolvimento

Siga estes passos para ter o ambiente rodando do zero em menos de 2 minutos:

### 1. Pré-requisitos
*   [Docker](https://www.docker.com/) e [Docker Compose](https://docs.docker.com/compose/)
*   [Go](https://golang.org/dl/) (instalado localmente para os geradores)

### 2. Configuração Inicial
Clone o repositório e configure as variáveis de ambiente:
```bash
cp .env.example .env
```

### 3. Subir a Infraestrutura
Inicie os serviços de banco de dados, cache e mensageria:
```bash
make infra-up
```

### 4. Rodar a Aplicação (Live Reload)
O comando abaixo instala o `Air` automaticamente e inicia o servidor com hot-reload:
```bash
make dev
```
A API estará disponível em `http://localhost:8888`.

---

## 📊 Observabilidade e Monitoramento

O boilerplate já vem configurado com ferramentas essenciais para monitoramento em produção:

*   **Métricas (Prometheus):** Disponíveis no endpoint `/metrics`. Expõe contadores de requisições, latência e status HTTP.
    *   URL: `http://localhost:8888/metrics`
*   **Logging Estruturado:** Utiliza o `uber-go/zap` para logs em JSON de alta performance.
    *   **TraceID:** Cada log inclui um `trace_id` único para rastrear requisições ponta-a-ponta através do middleware.
    *   **Contexto:** Suporte para logging com contexto para capturar metadados da requisição automaticamente.
*   **Health Check:** Endpoint simples para monitorar o status da aplicação.
    *   URL: `http://localhost:8888/health`

---

## 🛡️ Segurança e Proteção

A segurança é tratada como prioridade, com múltiplas camadas de proteção:

*   **Rate Limiting (Redis):** Proteção contra força bruta e DoS.
    *   **Inteligente:** Identifica usuários autenticados por ID e usuários anônimos por IP.
    *   **Headers:** Retorna `X-RateLimit-Limit`, `X-RateLimit-Remaining` e `X-RateLimit-Reset`.
    *   **Configurável:** Limites e janelas de tempo definidos via variáveis de ambiente.
*   **RBAC (Role-Based Access Control):** Controle de acesso granular baseado em permissões vinculadas a papéis (Roles).
    *   Middleware dedicado para checagem de permissões por endpoint.
*   **Gestão de Sessões:** Invalidação de tokens em tempo real via Redis (Logout global, troca de senha ou alteração de permissões).
*   **Auditoria Automática:** Registro de quem, quando e o que foi alterado em qualquer tabela do banco de dados através de GORM Hooks.

## 🛠️ Utilizando os Geradores

Aumente sua produtividade usando as ferramentas de automação inclusas:

### Criar um Novo Módulo CRUD
Gera Repository, Service, Handler, Routes, Models e **Testes de Integração** automaticamente:
```bash
make generate name=product
```

### Adicionar um Novo Driver de Storage
Gera a implementação, testes e integra o driver à Factory de Media:
```bash
make storage-driver name=cloudinary
```

---

## 🧪 Testes e Cobertura

A filosofia deste projeto é **100% de cobertura**. Não aceitamos código sem validação real.

*   **Executar Testes:**
    ```bash
    make test
    ```
*   **Gerar Relatório de Cobertura:**
    ```bash
    make coverage
    ```
*   **Visualizar Cobertura no Navegador:**
    ```bash
    make coverage-html
    ```

---

## 📜 Comandos Disponíveis (Makefile)

| Comando | Descrição |
| :--- | :--- |
| `make dev` | Inicia o servidor com Live Reload (Air) |
| `make test` | Executa todos os testes do projeto |
| `make coverage` | Gera resumo de cobertura de código no terminal |
| `make coverage-html`| Abre o relatório de cobertura detalhado no browser |
| `make swagger` | Atualiza a documentação OpenAPI/Swagger |
| `make generate name=X` | Gera um novo módulo CRUD completo |
| `make infra-up` | Sobe Postgres, Redis e RabbitMQ via Docker |
| `make infra-down` | Remove todos os containers de infraestrutura |

---

## 🚀 CI/CD (GitHub Actions)

O projeto conta com um pipeline de Integração Contínua (CI) robusto:

*   **Build:** Verifica se a aplicação compila corretamente em múltiplos ambientes.
*   **Testes Automatizados:** Executa toda a suíte de testes usando Testcontainers (Postgres/Redis reais).
*   **Coverage Guard:** O pipeline falha automaticamente se a cobertura de código for inferior a **100%**.
*   **Dependency Check:** Garante que o `go.mod` e `go.sum` estão sincronizados.

---

## 📖 Documentação da API

A documentação interativa via Swagger/OpenAPI está organizada por versão:

*   **Swagger UI:** `http://localhost:8888/v1/docs/index.html`
*   **JSON Spec:** `http://localhost:8888/v1/docs/doc.json`

---
Desenvolvido com ❤️ para ser a base definitiva de projetos Go de alta performance.
