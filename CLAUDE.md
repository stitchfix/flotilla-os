# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Overview

Flotilla is a self-service framework for defining and executing containerized jobs on AWS EKS (Elastic Kubernetes Service). It consists of:
- **Backend**: Go service with REST API for managing task definitions and runs
- **Frontend**: React/TypeScript UI for creating, running, and monitoring tasks
- **Workers**: Background processes that handle job lifecycle (submit, status, retry)
- **Execution**: Kubernetes-based job execution on EKS clusters

## Development Commands

### Backend (Go)

```bash
# Install dependencies
go get

# Build the binary
go build

# Run tests
go test -v ./...

# Run the service locally (requires PostgreSQL and AWS resources)
go run main.go conf/

# Using docker-compose (includes PostgreSQL)
docker-compose up -d
```

### Frontend (React/TypeScript)

```bash
cd ui/

# Install dependencies
npm install

# Start development server (runs on port 3001)
npm start

# Build for production
npm run build

# Run tests
npm test
```

### Database Migrations

The project uses Flyway for database migrations located in `.migrations/`:

```bash
# Baseline the database
flyway baseline -configFiles=./.migrations/dev.conf -user=flotilla -password=flotilla

# Run migrations
flyway migrate -configFiles=./.migrations/dev.conf -locations=filesystem:./.migrations/ -user=flotilla -password=flotilla
```

## Architecture

### Core Components

**State Management** (`state/`):
- `models.go`: Core data models (Definition, Run, Template, etc.)
- `pg_state_manager.go`: PostgreSQL-backed state persistence
- Constants: StatusQueued, StatusPending, StatusRunning, StatusStopped, StatusNeedsRetry

**Execution Engines** (`execution/engine/`):
- `engine.go`: Engine interface for job execution
- `eks_engine.go`: EKS-based execution engine for standard tasks
- `emr_engine.go`: EKS-based execution engine for Spark tasks (EMR-style)
- `dcm.go`: Dynamic cluster manager for routing jobs to appropriate clusters

**Workers** (`worker/`):
- `submit_worker.go`: Pulls jobs from queue and submits to Kubernetes
- `status_worker.go`: Polls job status and updates state
- `retry_worker.go`: Handles failed jobs that can be retried
- `events_worker.go`: Processes Kubernetes pod events
- Workers run as goroutines coordinated by `worker_manager.go`

**Services** (`services/`):
- `execution.go`: Orchestrates run creation and execution
- `definition.go`: Manages task definitions (CRUD operations)
- `template.go`: Manages reusable task templates
- `logs.go`: Fetches logs from CloudWatch or S3

**Queue Management** (`queue/`):
- `sqs_manager.go`: AWS SQS-based queue implementation
- Separate queues for EKS standard jobs and Spark jobs

**API Layer** (`flotilla/`):
- `app.go`: Application initialization and HTTP server setup
- `endpoints.go`: HTTP handlers for REST API
- `router.go`: Route definitions

### Job Lifecycle

1. **Definition**: User creates a task definition (image, command, resources, env vars)
2. **Execute**: User triggers a run from a definition, creating a Run object
3. **Queued**: Run is placed in SQS queue (StatusQueued)
4. **Submit**: Submit worker pulls from queue and creates Kubernetes Job (StatusPending)
5. **Running**: Pod starts executing (StatusRunning)
6. **Stopped**: Job completes with exit code (StatusStopped)
7. **Retry Logic**: If job fails due to infrastructure issues (null exit code), transitions to StatusNeedsRetry

### Configuration

Configuration uses Viper and is loaded from `conf/config.yml`. Key settings:
- `database_url`: PostgreSQL connection string
- `eks_clusters`: Comma-separated list of EKS cluster names
- `eks_log_namespace`: Kubernetes namespace for job logs
- `enabled_workers`: List of workers to run (submit, status, retry, events)
- `queue_namespace`: SQS queue prefix
- `execution_engine`: "eks" or "eks-spark"

Environment variables override config file values (e.g., `DATABASE_URL`, `FLOTILLA_MODE`).

### Frontend Architecture

React app with Redux Toolkit for state management:
- `state/`: Redux slices for definitions, runs, logs, templates
- `components/`: UI components organized by feature
- `helpers/`: Utility functions for API calls and data formatting
- `api.ts`: Axios-based API client configuration

The UI communicates with the backend API at `/api/v6/` endpoints.

## Key Patterns

### Engine Selection
Jobs are routed to either `EKSEngine` (standard containers) or `EKSSparkEngine` (Spark jobs) based on the definition's engine type.

### Resource Constraints
- CPU: 256-60000 millicores (256-94000 for GPU)
- Memory: 512-350000 MB (512-376000 for GPU)
- Ephemeral Storage: Up to 5000 MB
- Node lifecycle: "spot" (default) or "ondemand"

### Tracing
DataDog APM tracing is integrated throughout using `gopkg.in/DataDog/dd-trace-go.v1`.

### Error Handling
- Use `github.com/pkg/errors` for error wrapping with context
- Exceptions are tracked in `exceptions/errors.go`

## Testing

Tests use standard Go testing framework:
- Unit tests are co-located with source files (e.g., `config_test.go`)
- Mock implementations in `testutils/mocks.go`
- Run with `go test -v ./...`

## API

The REST API is versioned (currently v6) and follows these patterns:
- `POST /api/v6/task` - Create task definition
- `GET /api/v6/task` - List task definitions
- `PUT /api/v6/task/alias/:alias/execute` - Execute task by alias
- `GET /api/v6/history/:run_id` - Get run details
- `GET /api/v6/history/:run_id/logs` - Get run logs

Full API documentation: https://stitchfix.github.io/flotilla-os/api.html

## AWS Dependencies

Flotilla requires:
- **EKS**: Kubernetes clusters for job execution
- **PostgreSQL**: State storage (definitions, runs, templates)
- **SQS**: Job queuing
- **CloudWatch/S3**: Log storage
- **IAM**: Service account permissions for Kubernetes jobs

The service assumes AWS credentials are available via standard AWS SDK credential chain.
