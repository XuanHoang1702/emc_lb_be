---
type: project
created: 2026-05-25
updated: 2026-07-10
---

# Project Conventions

## Git Workflow
- Always create a new dedicated branch for major code changes.
- Branch name format should follow: `feature/[task-slug]` or `fix/[bug-slug]`.

## API Documentation
- Use Swaggo to generate API documentation from source code comments.
- Before committing, run `make swag` to update `src/docs`.
- Docs are accessible at `/docs/index.html`.

## Security & Rate Limiting
- **Public APIs** (e.g. login, OTP) MUST be protected by Rate Limiting to prevent DDoS/Brute-force.
- We use `github.com/go-redis/redis_rate/v10` implementing a Token Bucket algorithm via Redis.
- Rate Limiter is injected at the route group level for `publicGroup` inside `src/internal/routes/routes.go`.
- Currently set to 100 requests per 10 seconds per IP address.
