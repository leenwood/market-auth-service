You are helping build a production-ready Go microservice. Use the infrastructure
patterns described below as the baseline. Do not invent alternatives — replicate
these patterns exactly unless told otherwise.

## Project layout

Three binaries in cmd/: server (HTTP), worker (async), migrate (DB migrations).
All application code lives under internal/ in four layers enforced by go-arch-lint:
  - Foundation  (no internal deps): config, domain, platform/{logger,metrics,tracing}
  - Contracts   (→ Foundation):     port interfaces, service helpers, mappers
  - Infra       (→ Foundation + Contracts): storage, messaging, outbound HTTP
  - Application (→ all layers):     HTTP server wiring, worker/processor loop

Bootstrap lives in internal/app/service/ — it assembles and starts each binary.

## Config

Load all configuration from environment variables in a single internal.Load() call
that returns a *Config struct. Use typed sub-structs per concern:
HTTPConfig, PostgresConfig, KafkaConfig, ClickHouseConfig, LogConfig, OTelConfig, AppConfig.
Helper functions: getEnv(key, fallback), getEnvInt, getEnvBool, getEnvDuration,
getEnvStringSlice (comma-separated), requireEnv (returns error if missing).
Never use a third-party config library.

## Logger

Use stdlib log/slog. Constructor: logger.New(level, format string) *slog.Logger.
Levels: debug/warn/error/info (default). Format: "text" → TextHandler, anything else → JSONHandler.
Propagate request_id and trace_id through context.Context:
  logger.WithRequestID(ctx, id) / RequestIDFromContext(ctx)
  logger.WithTraceID(ctx, id)   / TraceIDFromContext(ctx)
  logger.FromContext(ctx, base) returns base.With(request_id, trace_id) fields attached.
All handler and service logs must go through logger.FromContext so every line carries
the correlation IDs automatically.

## Tracing (OpenTelemetry)

Package: internal/platform/tracing.
tracing.Init(ctx, Config) returns (ShutdownFunc, error).
If Config.Enabled == false, return a no-op shutdown immediately.
OTel resource includes ServiceName, process info, OS info (semconv v1.26.0).
Exporter: "otlp" → otlptracehttp (insecure, configurable endpoint); default → stdouttrace.
Sampler: AlwaysSample (change per service as needed).
Propagators: TraceContext + Baggage (W3C).
Register the tracer provider globally via otel.SetTracerProvider.

## Metrics (Prometheus)

Package: internal/platform/metrics.
Always use a non-global isolated Registry: prometheus.NewRegistry().
Register Go and process collectors on startup.
Standard metric set for HTTP services:
  platform_http_requests_total{method, path, status}       CounterVec
  platform_http_request_duration_seconds{method, path}     HistogramVec
    Buckets: .005 .01 .025 .05 .1 .25 .5 1 2.5 5
For outbound calls:
  platform_http_outbound_requests_total{target, method, status} CounterVec
For Kafka consumers:
  platform_queue_lag_messages{topic}                        GaugeVec
Expose metrics at GET /metrics using promhttp.HandlerFor(registry, ...) with
EnableOpenMetrics: true.

## HTTP server

Wire everything in apphttp.NewServer(cfg Config, deps Deps) *http.Server.
Standard routes every service must have:
  GET /health   → liveness  (returns 200 if app is up)
  GET /ready    → readiness (pings DB, returns 503 if not ready)
  GET /metrics  → Prometheus scrape endpoint
  /swagger/     → Swagger UI (httpSwagger.WrapHandler)
  GET /debug/pprof/* → only if cfg.PprofEnabled == true

Middleware chain (outermost → innermost, applied via middleware.Chain):
  1. otelhttp.NewHandler  — OTel server span (filter out /metrics and /health)
  2. middleware.Recover   — panic → 500 + log stack trace
  3. middleware.Logger    — log method/path/status/latency + Prometheus RED metrics
  4. middleware.RequestID — read X-Request-ID header or generate uuid, set on response
  5. middleware.MaxBodySize(1 MiB)

Logger middleware must inject trace_id into context from the OTel span before the
handler runs, so every log line inside handlers carries it.

HTTP server timeouts (all configurable via env):
  ReadTimeout: 15s, WriteTimeout: 15s, IdleTimeout: 60s.

## Graceful shutdown

In main(): signal.NotifyContext(ctx, SIGINT, SIGTERM).
The RunServer/RunWorker function blocks on select { srvErr | ctx.Done }.
On ctx.Done, create a new context.WithoutCancel + 30s timeout for shutdown.
Shutdown order: HTTP server → Kafka producer → ClickHouse → PostgreSQL → OTel flush.

## Infra bootstrap

Use an Infra struct that groups all shared deps:
  Cfg, Log, Metrics, DB (*postgres.DB), ChDB (*clickhouse.DB), Producer (*messaging.Producer)
  plus a private shutdownTracing func.
initInfra(ctx, serviceNameSuffix) (*Infra, error) initialises everything with a 10s
timeout context. On any init failure, close already-opened resources before returning error.
Infra.Shutdown(ctx) drains producer → closes ClickHouse → closes Postgres → flushes traces.

## Outbound HTTP client

internal/infra/outbound/httpclient — resilient client with:
  - Retry: up to MaxRetries (default 3) with exponential backoff + full jitter
      delay = rand(0, min(30s, 100ms * 2^attempt))
      Retryable statuses: 429, 502, 503, 504, any >=500
  - Circuit breaker (Closed → Open → HalfOpen → Closed):
      MaxFailures: 5 consecutive failures → Open
      OpenTimeout: 30s → probe with HalfOpen
      HalfOpenProbes: 2 successes → Closed
      Log state transitions (Warn level)
  - OTel client span per call with http.method, http.url, http.status_code attributes
  - Inject W3C trace context into outbound request headers
  - Record platform_http_outbound_requests_total metric on every attempt
Client.Do(ctx, method, path, body) returns (*Response, error).
Config.Target is a short label for metrics/logs (e.g. "notifications-api").

## Tooling

Task runner: go-task (Taskfile.yml).
Standard tasks: build, run, run:worker, test, test:integration, test:cover,
lint, arch, arch:graph, fmt, vet, docker:up, docker:down, docker:reset,
migrate:up, migrate:down, migrate:create, docs.

golangci-lint and go-arch-lint are declared as Go tool dependencies in go.mod
(tool block) and invoked via `go tool golangci-lint run ./...` and
`go tool go-arch-lint check`.

Unit tests: go test -race -count=1 -timeout=60s ./...
Integration tests: require -tags=integration, use testcontainers-go,
named *_integration_test.go, live alongside the code they test.
Tests must not mock databases — integration tests hit real containers.

Arch linter config in .go-arch-lint.yml — every new package must be added with
its allowed dependency set. The linter enforces the four-layer boundary.

## Migrations

Use goose. Migrations live in migrations/postgres/ and migrations/clickhouse/.
The cmd/migrate binary accepts -target=postgres|clickhouse.
Goose migrations come in pairs (Up/Down).

## Build

CGO_ENABLED=0, -trimpath, -ldflags="-s -w".
Binaries output to bin/.

## Import order

Enforced by goimports with local prefix github.com/<org>/<service>.
Groups: stdlib → external → internal (local).
