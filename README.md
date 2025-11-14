# 💰 Go Bank

> ⚠️ **Work in Progress** - A production-ready banking API demonstrating modern backend engineering with Go, clean architecture patterns, and comprehensive observability.

[![Go Version](https://img.shields.io/badge/Go-1.25.0-00ADD8?logo=go)](https://go.dev/)
[![Docker](https://img.shields.io/badge/Docker-Compose-2496ED?logo=docker)](https://www.docker.com/)
[![Architecture](https://img.shields.io/badge/Architecture-Clean%20Architecture-blue)]()
[![Status](https://img.shields.io/badge/Status-In%20Progress-yellow)]()

---

## 📋 Overview

**Go Bank** is a production-grade banking API built to showcase enterprise-level backend development practices. The project implements core banking operations (user management, accounts, authentication, transactions) with a focus on **scalability, observability, and production readiness**.

### Why This Project Stands Out

- **Multi-Role Architecture**: Separate deployable units for API server and background workers
- **Full Observability**: OpenTelemetry tracing, structured logging (ELK), and Prometheus metrics
- **Integration Testing**: Real database/cache containers via testcontainers

---

## 🏗️ Architecture

The application uses a **role-based execution model** allowing horizontal scaling:

- **API Server** (`--role=server`): Handles HTTP requests, runs multiple instances behind a load balancer
- **Background Workers** (`--role=worker-default`, `--role=worker-priority`): Process async tasks from separate queues
- **Shared Infrastructure**: PostgreSQL (database) and Redis (cache + task queue)

Each domain follows **Clean Architecture** with consistent layers: controllers, services, repositories, and models.

---

## ✨ Features

### Implemented
- ✅ **Authentication**: Sign up, login, JWT tokens with Redis-backed revocation
- ✅ **User Management**: Get/update profile, change password
- ✅ **Account Operations**: View accounts, account details, internal money transfers
- ✅ **Background Tasks**: Welcome emails, scheduled statements with retry logic (dummy without real email service)
- ✅ **Observability**: OpenTelemetry tracing, structured logging with correlation IDs, Prometheus metrics

### In Progress
- 🚧 **Password Reset Flow**: Forgot password, reset password with token verification
- 🚧 **Account Statements**: Generate and list PDF statements via async tasks
- 🚧 **Admin Operations**: Deposit/withdrawal endpoints via bank employees
- 🚧 **External Transfers**: IFSC-based transfers to external banks

---

## 🛠️ Tech Stack & Key Decisions

**Core Technologies**: Go · PostgreSQL · Redis · Docker

**Why These Choices**:
- **Bun ORM**: SQL-first approach over heavy abstractions - maintains control while providing safety
- **Asynq**: Redis-backed task queue with built-in retry logic and scheduling
- **Argon2id**: 2019 Password Hashing Competition winner - better than bcrypt for modern threats
- **OpenTelemetry**: Vendor-neutral observability - easy to swap APM providers
- **Testcontainers**: Real database/cache in tests - catches integration issues unit tests miss
- **Multi-environment configs**: YAML base + environment overrides - single source of truth

---

## 🚀 Running Locally

### Prerequisites
- Go 1.25.0+
- Docker & Docker Compose
- Make

### Quick Start

```bash
# Start infrastructure (PostgreSQL, Redis, ELK, Prometheus, etc.)
make up

# Run database migrations
make migrate-up

# Start API server (port 8080)
make run

# (Optional) Start background workers
make run-worker              # Default queue worker
make run-worker-priority     # Priority queue worker
```

### Available Services

| Service       | URL                          | Purpose                    |
|---------------|------------------------------|----------------------------|
| API           | http://localhost:8080        | REST API endpoints         |
| Metrics       | http://localhost:8080/metrics| Prometheus metrics         |
| Kibana        | http://localhost:5601        | Log visualization          |
| Grafana       | http://localhost:3000        | Metrics dashboards         |
| Prometheus    | http://localhost:9090        | Metrics collection         |

### Testing

```bash
# Run all tests
make test

# Run specific test package
make test-pkgs pkgs=./tests/authentication

# Run single test
make test-one pkg=./tests/healthz name=Test_CheckHealth

# Verbose test output
make test verbose=true
```

### Database Migrations

```bash
# Create new migration
make migrate-create name=migration_name

# Apply migrations
make migrate-up

# Rollback last migration
make migrate-down
```

---

## 🗂️ Project Structure

```
go-bank/
├── cmd/                    # Application entry points
│   ├── app.go             # Bootstrap logic (role-based startup)
│   ├── middleware/        # HTTP middlewares (auth, logging)
│   ├── router/            # Route configuration
│   ├── server/            # HTTP server lifecycle
│   └── worker/            # Background worker setup
├── internal/              # Domain logic (Clean Architecture)
│   ├── account/           # Account domain
│   ├── authentication/    # Auth domain (JWT, tokens)
│   ├── user/              # User domain
│   └── healthz/           # Health checks
├── pkg/                   # Shared infrastructure
│   ├── cache/             # Redis abstraction
│   ├── database/          # PostgreSQL client + helpers
│   ├── logger/            # Structured logging
│   ├── metrics/           # Prometheus metrics
│   ├── tasks/             # Task queue (Asynq)
│   ├── telemetry/         # OpenTelemetry setup
│   └── testutils/         # Test helpers
├── migrator/              # Database migrations (Goose)
├── tests/                 # Integration tests
├── config/                # Environment configs (YAML)
└── docker-compose.yaml    # Local development stack
```


## 🧪 Testing

**Integration-First Approach**: Uses real PostgreSQL and Redis containers via testcontainers - if it works in tests, it works in production. Fixture-based test data ensures consistency across test runs.

---

## 📝 Development Principles

- **Interface-Driven Design**: All major components define interfaces for testability
- **Context Propagation**: Request context flows through entire call chain with correlation IDs
- **Configuration as Code**: Environment-specific YAML configs with environment variable overrides

---

## 🤝 Contributing

This is a personal learning project, but feedback and suggestions are welcome!

---

## 📄 License

This project is open source and available for educational purposes.

---

## 👤 Author

**Syed Kamran Ahmed**  
[GitHub](https://github.com/skamranahmed) | [LinkedIn](https://linkedin.com/in/skamranahmed)

---

**Status**: 🚧 Work in Progress | **Last Updated**: November 2025