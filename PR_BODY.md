Implements remaining enterprise hardening items (follow-up to prior hardening work) per `ENTERPRISE_READINESS_REVIEW.md` sections 1.3, 2.2, 5.2, 5.3.

Changes
- Add dependency-free HTTP middleware package (`gomikrobot/internal/httpmw`):
  - Per-client-IP token-bucket rate limiting (configurable)
  - Panic recovery middleware for HTTP handlers
  - Request body size limiting (default 10 MiB)
- Add Kubernetes probe endpoints on both API and dashboard servers:
  - `GET /health` (liveness)
  - `GET /ready` (readiness)
- Add configurable graceful shutdown drain period using `http.Server.Shutdown()`

Config additions (`gateway`)
- `rateLimitRps` (default: 5)
- `rateLimitBurst` (default: 10)
- `maxBodyBytes` (default: 10485760)
- `shutdownTimeout` (default: 10s)

Notes
- Rate limiting is best-effort based on `X-Forwarded-For` / `X-Real-IP` / `RemoteAddr`.
