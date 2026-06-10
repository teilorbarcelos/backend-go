# Boilerplate Backend Go - Otimizações para Alta Performance, Escalabilidade e Demanda Extrema

Checklist de otimizações a serem aplicadas no boilerplate. Itens marcados com `[ ]` ainda não foram implementados.
Itens marcados com `[x]` já existem no projeto e servem como referência do estado atual.

---

## 0. Premissas Obrigatórias (aplicam-se a TODOS os itens abaixo)

Todo item deste checklist — bem como qualquer código novo, refatoração ou correção introduzida no projeto — **deve** atender simultaneamente às seguintes premissas. Nenhuma otimização é aceita se violar qualquer uma delas.

- **[SOLID]**
  - **S** — Single Responsibility: cada struct/função/módulo tem uma única razão para mudar. Middlewares, services e repositórios não acumulam responsabilidades de logging, validação e serialização.
  - **O** — Open/Closed: novas estratégias (cache, storage, messaging, hash) entram por composição de interfaces já existentes, sem modificar código de chamadas.
  - **L** — Liskov Substitution: qualquer implementação de `StorageProvider`, `Cache`, `EmailProvider`, `SessionStore`, `Repository` etc. deve ser substituível pela interface sem quebrar o consumidor.
  - **I** — Interface Segregation: interfaces pequenas e coesas. Nada de `*Repository` mega-interface — segregar por contexto (read/write, lookup, paginated).
  - **D** — Dependency Inversion: services dependem de **interfaces** (não de `*gorm.DB` ou `*redis.Client`); a injeção é feita em `main` (composition root).

- **[DRY]**
  - Lógica duplicada extraída para helpers/middleware compartilhados.
  - Configurações, error mapping, validações e mappings de DTO centralizados (evitar o switch repetido em handlers e o `gin.H{"error": ...}` espalhado).

- **[Clean Code]**
  - Nomes expressivos e sem abreviações crípticas.
  - Funções pequenas, com no máximo um nível de abstração por vez.
  - **Sem comentários óbvios** (ex.: `// Loop sobre users`) — apenas comentários que expliquem *porquê*, não *o quê*.
  - Sem side effects implícitos; retorno de erro explícito.
  - `gofmt` + `goimports` limpos, sem imports não usados.

- **[Testes — Cobertura > 95%]**
  - Toda otimização vem com **testes unitários** para o novo código E para o código refatorado.
  - Cobrir **caminhos felizes + caminhos de erro + edge cases** (timeouts, contexto cancelado, redis indisponível, sessão inválida, permissão negada, payload inválido).
  - Manter a meta de **> 95%** de cobertura em `statements`, `branches`, `functions` e `lines` (medida via `make coverage`).
  - Rodar `-race` no CI e localmente para detectar data races.
  - Benchmarks para hot paths (auth, rate limit, search paginated, session invalidation).

- **[Compliance E2E — `mage-backend-compliance`]**
  - O backend deve estar em conformidade com a suite E2E específica para Go (`mage test-go` ou equivalente) hospedada em `/home/teilor/MyProjects/mage-backend-compliance`.
  - Subir a suite com `make up` se não estiver rodando.
  - Quaisquer novos endpoints, middlewares, headers ou mudanças em payloads devem ter seus casos de teste correspondentes cobertos (ou ter a suite estendida).
  - Se a otimização envolver geração de PDF, garantir `react-pdf-service` rodando (`make dev` em `/home/teilor/MyProjects/react-pdf-service`).

- **[Quality Gate — SonarQube]**
  - O `Quality Gate` do SonarQube (porta 9000, `make up` em `/home/teilor/MyProjects/sonar-qube`) deve passar a cada PR/commit.
  - Zero `BUG`, zero `VULNERABILITY`, zero `CODE_SMELL` críticos.
  - Duplicação de código < 3% (DRY é premissa, não meta).
  - Cobertura reportada no Sonar deve bater com a meta > 95%.
  - Complexidade ciclomática por função < 15.
  - Sem `TODO`, `FIXME` ou código comentado em PRs.

- **[Verificação contínua]**
  Após cada item implementado, rodar localmente:
  ```bash
  go test -race -count=1 ./...
  make coverage
  golangci-lint run
  ```
  E validar:
  1. Suite de testes do projeto > 95% passando, > 95% cobertura.
  2. `mage-backend-compliance` rodando e suite Go passando.
  3. SonarQube disponível e Quality Gate passando.

---

## 1. Autenticação e Sessão (Crítico para Segurança e Performance)

- `[x]` JWT com HS256 (`pkg/security/jwt.go:49`)
- `[x]` Armazenamento de sessão no Redis (`internal/infra/session/session_manager.go`)
- `[x]` Middleware de autenticação via Bearer (`internal/middleware/auth.go`)
- `[x]` Hash bcrypt para senhas (`pkg/security/password.go`)
- `[x]` SHA-256 para hash de tokens em chaves Redis (`pkg/security/hash.go`)
- `[x]` Permissões embutidas no JWT (RBAC via claims) (`pkg/security/jwt.go:23`)
- `[x]` Refresh token com validade maior (7 dias) (`pkg/security/jwt.go:34`)

### 1.1. Session Versioning
- `[x]` Adicionar `session_version` (int) no model `Auth` — `internal/core/models/auth.go:28`
- `[x]` Embutir `session_version` no JWT (`JWTClaims`) — `pkg/security/jwt.go:22`
- `[x]` No middleware `Authenticate`, comparar `claims.SessionVersion` com valor armazenado no Redis via `GET session:ver:<userId>` — O(1), sem SCAN — `internal/middleware/auth.go:42`
- `[x]` Substituir `InvalidateUserSessions` por INCR atômico em Redis (`INCR session:ver:<userId>`) — `internal/infra/session/session_manager.go:43`
- `[x]` `generateToken` aceita `sessionVersion` — `pkg/security/jwt.go:38`
- `[x]` `prepareAuthResponse` lê `Auth.SessionVersion` e repassa ao `GenerateToken` — `internal/app/auth/service.go:219`
- `[x]` Após login, `SetSessionVersion` cacheia versão no Redis — `internal/app/auth/service.go:234`
- `[x]` No `UserService.Update/Delete/SetStatus`, bump da versão via `IncrementSessionVersion` — `internal/app/user/service.go:213`
- `[x]` No `RoleService.Update/Delete/SetStatus`, bump em massa via `BulkIncrementSessionVersion` — `internal/app/role/repository.go:90`

### 1.2. Hardening de autenticação
- `[ ]` Trocar `jwt.SigningMethodHS256` por **RS256/EdDSA** (chaves assimétricas) — permite verificação stateless por múltiplos serviços sem expor o segredo de assinatura.
- `[ ]` Implementar rotação de chaves (KIDs em `kid` do header JWKS).
- `[ ]` Reduzir duração do access token para 10-15min e usar refresh rotativo com detecção de reuso (revoga família inteira se o mesmo refresh for usado duas vezes).
- `[ ]` Adicionar `jti` (JWT ID) único por access token e armazenar no Redis como `SETEX session:access:<jti> <exp> 1` para revogação unitária O(1).
- `[ ]` Adicionar `aud` (audience) e `iss` (issuer) nas claims e validar no `ParseWithClaims` para evitar token reuse entre APIs.
- `[ ]` Migrar `bcrypt.DefaultCost` para **argon2id** (`golang.org/x/crypto/argon2`) — bcrypt é CPU-intensive e bloqueia a goroutine; argon2id com memória fixa é mais resistente a GPU e tunable.
- `[ ]` Adicionar limitador de tentativas de login (sliding window no Redis) + lockout temporário por `email` e por `IP`.
- `[ ]` Implementar 2FA (TOTP) opcional para perfis sensíveis (administrator).

### 1.3. Autorização (RBAC)
- `[ ]` Compilar permissões em um `map[string]uint8` de bitset por usuário no momento do login e cachear no Redis (`session:perms:<userId>`) para lookup O(1) no `rbac.go` em vez de loop linear.
- `[ ]` Substituir switch de strings no `hasPermission` (`internal/middleware/rbac.go:11`) por lookup em mapa de bits.
- `[ ]` Avaliar Casbin/OPA apenas se a matriz de permissões crescer (não vale a complexidade hoje).

---

## 2. Banco de Dados (PostgreSQL + GORM)

- `[x]` Connection pool configurado (`pkg/database/db.go:59` — max 50, idle 5, lifetime 30m)
- `[x]` Migrations com golang-migrate em produção (`pkg/database/db.go:64`)
- `[x]` AutoMigrate em dev (`pkg/database/db.go:68`)
- `[x]` Soft delete via `is_deleted` (`internal/core/repository/base_repository.go:44`)
- `[x]` `COUNT(*)` separado do `FIND()` para paginação correta (`internal/core/repository/base_repository.go:134`)
- `[x]` Query builder dinâmico com filtros/search/ordenação (`pkg/database/query.go`)

### 2.1. Otimizações de pool e conexão
- `[ ]` Parametrizar `MaxOpenConns`, `MaxIdleConns`, `ConnMaxIdleTime` e `ConnMaxLifetime` via env (em vez de hard-coded em `pkg/database/db.go:59-61`). Default sugerido para alta concorrência: `MaxOpenConns = 2 * NumCPU`, `MaxIdleConns = NumCPU`, `ConnMaxIdleTime = 5m`, `ConnMaxLifetime = 30m`.
- `[ ]` Adicionar `SetConnMaxIdleTime` (GORM expõe via `sqlDB`).
- `[ ]` Habilitar `pgx` puro no lugar de `lib/pq` (`gorm.io/driver/postgres` aceita `pgx` via DSN) — melhor performance e suporte a `pgxpool` se quisermos trocar o driver.
- `[ ]` Configurar `statement_timeout` e `idle_in_transaction_session_timeout` no `AfterConnect` hook do pool para evitar queries/zombies que esgotam conexões.

### 2.2. Otimizações de query
- `[ ]` **Eliminar `Preload` em loops N+1**: hoje `user/handler.go:152` faz `Preload("Auth", "Role")` na listagem — usar `Joins` com `SELECT` explícito de colunas e separar contagem de busca.
- `[ ]` Adicionar `EXPLAIN`-aware logging em dev (gorm `logger` em modo `Warn` para queries >100ms) — já temos logger; falta ativar threshold.
- `[ ]` Substituir `Count(&total)` antes do `Find` por **count paralelo via goroutine** (já sugerido) ou, melhor, **window function** `COUNT(*) OVER()` retornando total na própria query quando o LIMIT não for absurdamente grande.
- `[ ]` Padronizar uso de `SELECT <colunas>` em vez de `SELECT *` para evitar I/O desnecessário em tabelas largas.
- `[ ]` Adicionar índices sugeridos automaticamente: revisar `EXPLAIN` de `SearchPaginated` (filtros `is_deleted`, `active`, FKs de `id_role`, `id_auth`, `id_feature`).
- `[x]` Habilitar `PrepareStmt: true` no `gorm.Config` — cache de prepared statements — `pkg/database/db.go:44`
- `[x]` Adicionar `NowFunc` no `gorm.Config` para consistência de timestamps UTC — `pkg/database/db.go:45`
- `[x]` `SetConnMaxIdleTime(5m)` para gerenciar conexões ociosas — `pkg/database/db.go:62`
- `[x]` Aumentar `MaxIdleConns` de 5 para 10 — `pkg/database/db.go:60`
- `[ ]` Mover `applyFiltersLogic` para uso de placeholders nomeados onde possível, evitando `fmt.Sprintf` em hot path.

### 2.3. Otimizações de audit hook
- `[ ]` O hook de `audit:update` (`internal/core/audit/hooks.go:48`) faz um `SELECT` extra antes do `UPDATE` para capturar `oldValues`. Mover para **async batch writer** (canal + flush periódico a cada 1s ou 100 registros) com persistência em batch via `INSERT ... VALUES (...), (...), ...`.
- `[ ]` Em alta concorrência, enfileirar logs no buffer channel com backpressure e usar `COPY` ou `multi-row INSERT` (GORM `CreateInBatches`).
- `[ ]` Adicionar opção de desabilitar audit por rota/modelo via tag `gorm:"audit:false"`.

---

## 3. Cache e Redis

- `[x]` Redis client singleton (`pkg/cache/redis.go:12`)
- `[x]` Rate limit via `Incr + TTL` em pipeline (`internal/middleware/ratelimit.go:44`)
- `[x]` Sessão e permissões em Redis (`internal/infra/session/session_manager.go`)

### 3.1. Otimizações de cache
- `[ ]` Trocar `cache.RedisClient` (global) por **injeção de dependência** com interface `Cache` para facilitar mock e swap por cluster (Redis Cluster, Dragonfly, KeyDB).
- `[ ]` Configurar `PoolSize`, `MinIdleConns`, `ReadTimeout`, `WriteTimeout`, `DialTimeout` no `redis.Options` (atualmente usa defaults — em produção isso é fonte de stalls).
- `[ ]` Habilitar `ContextTimeout` curto (ex.: 200ms) em todas as chamadas — atual passa `c.Request.Context()` sem deadline definido, herdando timeout do cliente HTTP.
- `[ ]` Implementar **cache-aside genérico** (`pkg/cache/aside.go`) com singleflight (`golang.org/x/sync/singleflight`) para evitar thundering herd em lookups de user/role/features.
- `[ ]` Implementar **stampede protection** nas buscas pesadas: ao expirar chave, apenas uma goroutine regenera; as demais esperam.
- `[ ]` Adicionar `Redis Sentinel` ou `Redis Cluster` mode (configurável por env) com reconexão automática.
- `[ ]` Usar `UNLINK` em vez de `DEL` para deleção de chaves grandes (`session_manager.go:69`) — `UNLINK` é non-blocking.
- `[ ]` Substituir `Scan + Del` no `InvalidateUserSessions` por bump de versão (ver 1.1) ou por `Set` em índice reverso (`SET sessions:user:<id> <set de chaves>` + `EXPIRE`).
- `[ ]` Ativar `EnablePipelineMultiplex` (pipelining de comandos concorrentes) onde aplicável.

---

## 4. Middlewares e Request Lifecycle

- `[x]` Logger com request ID e latência (`internal/middleware/logger.go`)
- `[x]` ErrorLogger persistindo em DB async (`internal/middleware/error_logger.go`)
- `[x]` Métricas Prometheus (`internal/middleware/metrics.go`)
- `[x]` CORS (`internal/middleware/cors.go`)
- `[x]` Rate limit (`internal/middleware/ratelimit.go`)
- `[x]` Recovery via `gin.Recovery()` (`cmd/api/main.go:63`)

### 4.1. Otimizações de pipeline
- `[ ]` Reordenar middlewares do mais barato ao mais caro: `Recovery -> RequestID -> Metrics -> CORS -> RateLimit -> Auth -> Logger -> ErrorLogger`. Atual ordem em `cmd/api/main.go:63-68` põe Logger e ErrorLogger antes do Auth — métricas e rate limit não contabilizam tentativas não autenticadas corretamente.
- `[ ]` Implementar middleware de **compressão** (`gin-contrib/gzip` ou `klauspost/compress`) com threshold (ex.: 1KB) e níveis por tipo MIME.
- `[ ]` Adicionar `Cache-Control` e `ETag` em respostas de leitura para permitir 304 sem body.
- `[ ]` Implementar **request body size limit** (proteção contra DoS) — `c.Request.Body = http.MaxBytesReader(w, body, N)`.
- `[ ]` **Time budget middleware**: deadline máximo por request (ex.: 5s) propagado via context para GORM/Redis, retornando 504 se exceder.
- `[ ]` Substituir `gin.New()` por `gin.New()` customizado com `Engine.NoRoute` e `NoMethod` retornando JSON tipado (em vez de HTML default).
- `[ ]` Adicionar **circuit breaker** (`sony/gobreaker`) em chamadas a serviços externos (PDF service, RabbitMQ publish) — `internal/infra/pdf/client.go` e `pkg/messaging/rabbitmq.go`.
- `[ ]` Adicionar **timeout** em `http.Client` do PDF service (`internal/infra/pdf/client.go`) e `rabbitmq.Publish` — hoje `Publish` usa 5s hard-coded; PDF service não tem timeout explícito.

### 4.2. Logger
- `[ ]` Trocar `zap.NewProductionConfig()` por encoder JSON com sampling (rate limit de logs repetidos) em alta carga (`zap.NewProductionConfigWithOptions` com `Sampling`).
- `[ ]` Ativar `AddStacktrace` somente em `ErrorLevel`.
- `[ ]` Async sink (BufferedWriteSyncer) para evitar bloqueio de I/O no hot path.
- `[ ]` Propagar `requestId` via `context` (typed key em vez de string) para evitar colisão de nomes — `logger.WithContext` (`pkg/logger/logger.go:60`) usa chave string `"requestId"`.
- `[ ]` Estruturar campos do Logger (`zap.Field` alocando menos que `fmt.Sprintf` em `error_logger.go:31,38`).

### 4.3. Rate limit
- `[ ]` Migrar para **sliding window log** ou **token bucket** (mais justo que fixed window com `INCR`).
- `[x]` Usar **Redis Lua script** atômico em vez de pipeline+`Expire` separado — `internal/middleware/ratelimit.go:15-25`
- `[ ]` Suportar rate limit por **rota + usuário** (atualmente é global por usuário/IP).
- `[ ]` Adicionar burst control (e.g., 200 req instantâneo, depois 100/min sustentado).

### 4.4. Métricas
- `[ ]` Reduzir cardinalidade do label `path` em `metrics.go:38-39` — `c.FullPath()` é bounded (rotas registradas), mas rotas com path param (e.g., `/user/:id`) geram a mesma string, ok. Validar que nenhum `path` dinâmico está vazando.
- `[ ]` Adicionar métricas de: latência de DB, latência de Redis, tamanho de payload HTTP, GC pause.
- `[ ]` Adicionar tracing distribuído com OpenTelemetry (já temos `go.opentelemetry.io/otel` como dependência indireta em `go.mod:163-167`).

---

## 5. Modelos e Domain

- `[x]` `BaseModel` com soft delete (`internal/core/models/base.go`)
- `[x]` Anonimização LGPD no delete de usuário (`internal/app/user/service.go:180`)

### 5.1. Otimizações
- `[ ]` Adicionar `gorm:"index:..."` em colunas usadas em filtros frequentes (`email`, `is_deleted`, `active`, `id_role`, `id_auth`).
- `[ ]` Adicionar **composite indexes** para queries comuns: `(is_deleted, active, created_at DESC)` em quase todas as tabelas de listagem.
- `[ ]` Adicionar `UpdatedAt` em `RoleFeature` se ainda não tem (verificar).
- `[ ]` Substituir `*string` (phone, document, avatar) por tipos mais expressivos quando o domínio permitir (e.g., `phonenumber`, mas Go ainda não tem padrão — manter ponteiro é correto, mas garantir que validação existe no DTO).
- `[ ]` Validar que `Role` preload em `user/handler.go:152` está usando `Joins` em vez de `Preload` na listagem (N+1 silencioso).

---

## 6. Configuração e Bootstrap

- `[x]` Viper para carregar env e `.env` (`pkg/config/env.go`)
- `[x]` `viper.Unmarshal` em `AppConfig` global
- `[x]` Defaults seguros para todas as envs

### 6.1. Otimizações
- `[ ]` Validar `JWT_SECRET` em produção (length >= 32 bytes, falhar boot se fraco).
- `[ ]` Validar `RateLimitMax`/`RateLimitWindow` em produção (atualmente pode estar zerado e rate limit vira `>= 0` sempre true).
- `[ ]` Substituir `viper.Unmarshal` por **struct tags com validação** em startup (`go-playground/validator` já é dependência).
- `[ ]` Carregar `.env` somente em dev — em prod ler apenas env vars reais, evitando I/O de FS.
- `[ ]` Carregar config uma única vez e injetar (DI) em vez de variável global `AppConfig`.
- `[ ]` Suportar hot-reload de config sem restart (viper `WatchConfig`).

---

## 7. Mensageria (RabbitMQ)

- `[x]` Wrapper com `PublishWithContext` (`pkg/messaging/rabbitmq.go:64`)
- `[x]` `QueueDeclare` com `durable=true` (`pkg/messaging/rabbitmq.go:74`)

### 7.1. Otimizações
- `[ ]` Trocar canal global por **pool de canais** (um por goroutine worker) — RabbitMQ recomenda 1 canal por thread para alta concorrência.
- `[ ]` Configurar `prefetchCount` (QoS) no consumer — não temos consumer ainda, mas antecipar.
- `[ ]` Habilitar **Publisher Confirms** (`channel.Confirm(false)`) com ack assíncrono para garantir entrega em prod.
- `[ ]` Adicionar **retry com backoff exponencial** para publishes que falham.
- `[ ]` Trocar `json.Marshal` por `goccy/go-json` ou `bytedance/sonic` (já é dependência via Gin) para reduzir alloc.
- `[ ]` Adicionar **dead-letter exchange** para mensagens não processáveis.
- `[ ]` Validar reconexão automática em caso de queda (hoje `main.go:122-124` só fecha; não tenta reconectar).

---

## 8. Storage (S3 / Local / etc)

- `[x]` Interface `StorageProvider` (`pkg/storage/interface.go`)
- `[x]` Factory pattern (`pkg/storage/factory.go`)
- `[ ]` **Implementar drivers**: `local` (FS), `s3` (MinIO/AWS), `gcs`, `azure` — factory atualmente retorna erro para qualquer driver (`pkg/storage/factory.go:7-9`).
- `[ ]` Implementar **presigned URLs** com TTL para download direto (evita proxy pelo backend).
- `[ ]` Adicionar **streaming upload** (multipart) para arquivos grandes.
- `[ ]` Adicionar validação de MIME type no upload (não confiar em Content-Type do cliente).

---

## 9. Validação

- `[x]` `go-playground/validator/v10` (`pkg/validator/validator.go`)
- `[x]` Tags `binding` nos DTOs (`internal/app/user/handler.go:80-93`)
- `[ ]` Centralizar mensagens de erro de validação em PT-BR (i18n) — atualmente Gin retorna mensagens default em inglês.
- `[ ]` Validar DTOs antes de chegar no service (hoje dependemos de `c.ShouldBindJSON` no handler — ok, mas extrair para um helper evita duplicação).
- `[ ]` Validar tamanhos máximos de campos string (`binding:"max=255"`) para evitar abuse.

---

## 10. HTTP Server (Gin)

- `[x]` Gin com recovery (`cmd/api/main.go:62`)
- `[x]` Swagger UI em `/v1/docs` (`cmd/api/main.go:86`)
- `[x]` Health check `/health` (`cmd/api/main.go:70`)

### 10.1. Otimizações
- `[x]` Trocar `http.Server` default por configuração explícita: `ReadHeaderTimeout`, `ReadTimeout`, `WriteTimeout`, `IdleTimeout`, `MaxHeaderBytes`. **Crítico para segurança** — sem `ReadHeaderTimeout` há vetor de slowloris. — `cmd/api/main.go:97-103`
- `[ ]` Ativar `http2` (Gin suporta via `RunTLS` ou server customizado).
- `[ ]` Configurar `GracefulShutdown` com tempo >5s se houver workers longos (atualmente 5s em `main.go:114`).
- `[ ]` Pool de `context.Context` cancelados (evitar leak em conexões que demoram).
- `[ ]` Desabilitar `Swagger` em produção ou proteger com auth (atualmente público em `cmd/api/main.go:83-86`).
- `[ ]` Sanitizar respostas de erro em produção — não vazar stack traces ou detalhes internos.
- `[ ]` Validar `TrustedProxies` para `c.ClientIP()` funcionar corretamente atrás de load balancer.

---

## 11. Email

- `[x]` Interface `Provider` (`pkg/email/email.go:15`)
- `[x]` Mock provider para dev (`pkg/email/email.go:21`)

### 11.1. Otimizações
- `[ ]` Implementar provider real (SES, SendGrid, SMTP).
- `[ ]` Enfileirar emails no RabbitMQ em vez de enviar no request hot path (atualmente `auth.service.go:156` chama `emailProvider.SendEmail` síncrono).
- `[ ]] Adicionar templates (text + html) com i18n.
- `[ ]` Adicionar retry e tracking de entrega.

---

## 12. Testes e Qualidade

- `[x]` Testcontainers para Postgres e Redis (`pkg/testutil/containers.go`)
- `[x]` Miniredis para testes rápidos (`pkg/cache/redis.go:21`)
- `[x]` 100% de cobertura como meta (`Makefile`)

### 12.1. Otimizações
- `[ ]` Adicionar **benchmarks** para hot paths: `BaseRepository.SearchPaginated`, `session.InvalidateUserSessions`, `ratelimit`, `auth.Login`.
- `[ ]] Adicionar testes de **concorrência** (`go test -race`) e fuzzing de inputs (SQL injection, JSON malformado).
- `[ ]` Adicionar **contract tests** para integrações externas (PDF service, RabbitMQ).
- `[ ]` Rodar `go test -race` no CI.
- `[ ]` Rodar `golangci-lint` com perfil restritivo (`errcheck`, `gosec`, `gocritic`, `gocyclo`).
- `[ ]] Adicionar testes de carga (`k6`, `vegeta`) e SLOs documentados (p99 < 200ms em login, etc.).
- `[ ]` Validar com `go vet`, `staticcheck`, `gopls check`.

---

## 13. Segurança

- `[ ]` Adicionar **CSRF protection** para rotas com cookie session (não se aplica se usarmos apenas Bearer, mas validar).
- `[ ]] Adicionar headers de segurança: `X-Content-Type-Options: nosniff`, `X-Frame-Options: DENY`, `Strict-Transport-Security`, `Content-Security-Policy`, `Referrer-Policy`.
- `[ ]` Validar input de **path/query** params contra injection (especialmente `searchFields` em `query_parser.go:53-71`).
- `[ ]` Sanitizar logs para não gravar PII (senhas, tokens, CPF).
- `[ ]` Adicionar **secret scanning** no CI (gitleaks/trufflehog).
- `[ ]` Forçar HTTPS em produção (`Strict-Transport-Security` + redirect).
- `[ ]] Implementar **rate limit adaptativo** (captcha/throttling após N falhas de login).
- `[ ]` Implementar **device fingerprinting** e alerta de novo dispositivo.
- `[ ]] Adicionar **audit log imutável** (append-only, com hash chain) para conformidade.

---

## 14. Observabilidade

- `[x]` Prometheus metrics básico (`internal/middleware/metrics.go`)
- `[x]` Request ID propagation (`internal/middleware/logger.go:21`)
- `[x]` Health check (`cmd/api/main.go:70`)

### 14.1. Otimizações
- `[ ]` Adicionar **readiness vs liveness** distintos (readiness checa DB+Redis+RabbitMQ; liveness só o processo).
- `[ ]` Adicionar tracing distribuído (OpenTelemetry) — dependência já presente em `go.mod:163-167`.
- `[ ]` Exportar métricas customizadas de domínio: logins/min, sessions ativas, error rate por rota.
- `[ ]` Configurar `pprof` endpoint protegido por auth para análise de CPU/memória.
- `[ ]] Dashboards Grafana pré-configurados (templates em `infra/grafana/`).
- `[ ]] Alertas Prometheus (error rate > 1%, p99 > 500ms, DB pool > 80%).

---

## 15. CI/CD e DevOps

- `[x]` Dockerfile presente (`Dockerfile`)
- `[x]` docker-compose para infra e metrics
- `[x]` GitHub Actions (`.github/`)

### 15.1. Otimizações
- `[ ]` Multi-stage build no `Dockerfile` para reduzir tamanho da imagem final.
- `[ ]` Build com `-trimpath -ldflags="-s -w"` para binário menor e sem info de path.
- `[ ]` Imagem distroless ou scratch para produção.
- `[ ]` Rodar como **usuário não-root** no container.
- `[ ]` Adicionar `HEALTHCHECK` no Dockerfile.
- `[ ]` Adicionar `Makefile` targets: `make build`, `make run-prod`, `make lint`, `make security-scan`.
- `[ ]` Pipeline de CI: lint -> test -> build -> push -> deploy.
- `[ ]` Versionamento semântico via git tags + `internal/version` package.

---

## 16. Resiliência

- `[ ]` **Health check de dependências** no `/health` (DB ping, Redis ping, RabbitMQ ping).
- `[ ]` **Graceful shutdown** deve drenar conexões HTTP ativas antes de fechar pools (atualmente `main.go:114-119` chama `srv.Shutdown` mas não espera conexões DB/Redis em uso).
- `[ ]` **Connection retry** com backoff exponencial em `ConnectDB`, `ConnectRedis`, `ConnectRabbitMQ` (atualmente `logFatalf` em qualquer falha).
- `[ ]` Implementar **bulkhead pattern** (pools separados para queries críticas vs reporting).
- `[ ]` Documentar **runbook** de incidentes.

---

## 17. Documentação e Developer Experience

- `[x]` README presente
- `[x]` Swagger gerado automaticamente
- `[ ]` Documentar **variáveis de ambiente** em tabela no README (atualmente só `.env.example`).
- `[ ]` ADRs (Architecture Decision Records) em `docs/adr/`.
- `[ ]` Diagrama de sequência para fluxos críticos (login, refresh, CRUD).
- `[ ]` Postman/Insomnia collection exportada de `/docs/swagger.json`.
- `[ ]` Guia de contribuição (CONTRIBUTING.md).
- `[ ]` Changelog versionado.

---

## 18. Performance - Quick Wins (ordem de prioridade)

1. `[ ]` **Session versioning + bump O(1)** (item 1.1) — substitui SCAN/DEL em produção, ganho enorme em invalidação de sessão.
2. `[x]` **`PrepareStmt` + `NowFunc` + pool tuning no GORM** (item 2.2) — reduz parse e padroniza timestamps.
3. `[ ]` **HTTP server timeouts explícitos** (item 10.1) — fecha vetor slowloris e melhora UX.
4. `[x]` **Rate limit com Lua script atômico** (item 4.3) — elimina race entre INCR e EXPIRE.
5. `[ ]` **Async batch audit writer** (item 2.3) — move 1 query extra de cada UPDATE para batch assíncrono.
6. `[ ]` **Cache-aside + singleflight** (item 3.1) — protege backend de stampede em dados quentes.
7. `[ ]` **Compressão HTTP + ETag** (item 4.1) — reduz banda em 60-80% para JSON.
8. `[ ]` **Permissões em bitset cacheado** (item 1.3) — RBAC O(1) por request.
9. `[ ]` **OpenTelemetry tracing** (item 14.1) — visibilidade fim-a-fim.
10. `[ ]` **Argon2id no lugar de bcrypt** (item 1.2) — não bloqueia goroutines em CPU-intensive hashing.

---

## 19. Estado Atual vs Target

| Categoria                    | Atual | Target Pós-Otimização |
|------------------------------|-------|----------------------|
| Auth: invalidação de sessão  | O(n) SCAN+DEL | O(1) version bump |
| Hash de senha                | bcrypt DefaultCost | argon2id com tune |
| DB: pool de conexões         | hard-coded | env-driven + tuning |
| DB: queries                  | Preload N+1 latente | Joins + select colunas |
| DB: audit hook               | síncrono + SELECT extra | async batch writer |
| Rate limit                   | pipeline+Expire (race) | Lua atômico |
| HTTP server                  | defaults | timeouts explícitos |
| Cache                        | global client | DI + singleflight |
| Mensageria                   | canal global | pool + confirms |
| Storage                      | factory sem driver | s3/local/gcs prontos |
| Observabilidade              | métricas + logs | + tracing + pprof + alertas |
| Segurança                    | Bearer + bcrypt | RS256 + argon2id + 2FA |
