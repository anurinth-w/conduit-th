# Conduit-TH

Field management platform for waterworks contractors in Thailand.

## Tech Stack

| Layer | Technology |
|---|---|
| Backend | Go 1.23 (monorepo) |
| Database | PostgreSQL 16 + Redis 7 |
| Storage | MinIO (S3-compatible) |
| Infrastructure | Docker · Terraform · VPS |
| CI/CD | GitHub Actions |

## Quick Start

```bash
cp .env.example .env
make up
export DATABASE_URL=postgres://conduit:conduit_dev@localhost:5432/conduit_dev?sslmode=disable
make migrate
make test
```

## Services

| Service | Port | Description |
|---|---|---|
| gateway | 8000 | API Gateway · routing · auth middleware |
| auth | 8001 | JWT · refresh tokens |
| user | 8002 | Users · roles · company memberships |
| job | 8003 | Jobs · assignments · status machine |
| material | 8004 | Materials · pricing |
| media | 8005 | Photo upload · MinIO · Drive backup |
| document | 8006 | PDF generation · templates |
| notify | 8007 | Email · Line push notifications |
| report | 8008 | Analytics · CSV/PDF export |
| config | 8009 | Company config · job code formats · triggers |
| line-webhook | 8010 | Line OA webhook handler |
| ai-worker | — | AI/OCR background worker |

## Local Tools

| Tool | URL |
|---|---|
| Adminer (DB UI) | http://localhost:8080 |
| MinIO Console | http://localhost:9001 |
