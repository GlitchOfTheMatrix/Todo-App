# TaskFlow – Todo App (Scaled for 50k Users)

A full-stack task management application built with **Go** (backend) and **React + TypeScript** (frontend), architected to handle **50,000 concurrent users** reliably.

---

## Architecture Overview

```
                    ┌─────────────────┐
                    │   React + Vite  │
                    │   (Frontend)    │
                    └────────┬────────┘
                             │ HTTP / JSON
                    ┌────────▼────────┐
                    │   Nginx / LB    │ ← optional reverse proxy
                    └────────┬────────┘
                             │
                    ┌────────▼────────┐
                    │  Go HTTP Server │
                    │  (net/http)     │
                    │                 │
                    │ Middleware Stack│
                    │ ├─ CORS        │
                    │ ├─ Logger      │
                    │ ├─ Rate Limit  │
                    │ ├─ Body Limit  │
                    │ ├─ Req Timeout │
                    │ └─ Auth (JWT)  │
                    └────────┬────────┘
                             │
                    ┌────────▼────────┐
                    │   PostgreSQL    │
                    │   (pgxpool)     │
                    │  100 conn pool  │
                    └─────────────────┘
```

---

## Tech Stack

| Layer       | Technology                     |
|-------------|-------------------------------|
| Frontend    | React 19, TypeScript, Vite     |
| Backend     | Go 1.26, net/http (stdlib)     |
| Database    | PostgreSQL with pgx/v5 pool    |
| Auth        | JWT (HS256) via golang-jwt     |
| Password    | bcrypt                         |

---

## Scaling Changes Made

The following modifications were made to support 50,000 concurrent users:

### 1. Database Connection Pool Tuning

**File:** `backend/internal/database/database.go`

The database layer now accepts configurable pool parameters:

| Parameter           | Default   | Purpose                                              |
|---------------------|-----------|------------------------------------------------------|
| `DB_MAX_CONNS`      | `100`     | Maximum connections in the pool                      |
| `DB_MIN_CONNS`      | `10`      | Minimum idle connections kept warm                   |
| `DB_MAX_CONN_LIFETIME` | `30m`  | Recycles connections to prevent stale TCP sessions   |
| `DB_MAX_CONN_IDLE_TIME` | `5m`  | Closes idle connections to free DB resources          |
| `DB_HEALTH_CHECK_PERIOD` | `30s` | Periodic health checks to detect dead connections    |

**Why:** With default settings, pgxpool creates only 4 connections. Under 50k users, connection contention causes request queuing and timeouts. A pool of 100 connections (matching PostgreSQL's default `max_connections`) provides the headroom needed.

### 2. Rate Limiting (Per-IP Token Bucket)

**File:** `backend/internal/middleware/ratelimit.go`

A zero-dependency, in-memory token-bucket rate limiter:

| Parameter        | Default | Purpose                                |
|------------------|---------|----------------------------------------|
| `RATE_LIMIT_RPS` | `100`   | Requests per second per IP             |
| `RATE_LIMIT_BURST`| `200`  | Burst capacity for traffic spikes      |

- Extracts client IP from `X-Forwarded-For` → `X-Real-IP` → `RemoteAddr`
- Returns `429 Too Many Requests` with `Retry-After` header
- Automatic stale-entry cleanup every 3 minutes
- Thread-safe via `sync.Mutex`

**Why:** Without rate limiting, a single abusive client can saturate the server, causing latency for all 50k users.

### 3. Request Body Size Limiting

**File:** `backend/internal/middleware/limits.go`

| Parameter              | Default   | Purpose                                |
|------------------------|-----------|----------------------------------------|
| `MAX_REQUEST_BODY_SIZE`| `1048576` | Maximum body size in bytes (1 MB)      |

Uses Go's built-in `http.MaxBytesReader` to reject oversized payloads before they consume memory.

**Why:** Prevents memory exhaustion attacks. At 50k users, even a handful of large payloads can spike memory usage.

### 4. Request Timeout

**File:** `backend/internal/middleware/limits.go`

Every request gets a 30-second context deadline. If a database query or downstream call hangs, the context cancels and the goroutine is freed.

**Why:** Prevents goroutine leaks. Hung requests accumulate goroutines, each consuming ~8KB stack space. At 50k users, this leads to OOM kills.

### 5. Server-Side Pagination

**Files:**
- `backend/internal/repository/task_repository.go`
- `backend/internal/handler/task_handler.go`
- `backend/internal/models/task.go`
- `frontend/src/api/index.ts`
- `frontend/src/pages/DashboardPage.tsx`

Previously, `GET /api/tasks` returned **all** tasks for a user in a single response. Now it supports:

```
GET /api/tasks?limit=50&offset=0
```

Response shape:
```json
{
  "data": [...],
  "total": 1253,
  "limit": 50,
  "offset": 0,
  "has_more": true
}
```

- Default page size: 50
- Maximum page size: 100 (server-enforced)
- Frontend uses a "Load More" button for progressive loading

**Why:** A power user with 10,000 tasks would generate a multi-megabyte JSON response without pagination, multiplied by concurrent requests = network and memory exhaustion.

### 6. Database Indexes

**File:** `backend/migrations/3_add_indexes.up.sql`

```sql
CREATE INDEX idx_todos_user_id ON todos (user_id);
CREATE INDEX idx_todos_user_id_status ON todos (user_id, status);
CREATE INDEX idx_todos_user_id_created ON todos (user_id, created_at DESC);
```

**Why:** Without indexes, `SELECT ... WHERE user_id = $1 ORDER BY created_at DESC` performs a sequential scan. With 50k users × ~50 tasks each = 2.5M rows, this becomes O(n) per request instead of O(log n).

### 7. Structured JSON Logging

**File:** `backend/internal/middleware/logger.go`

Log output changed from plain text to structured JSON:

```json
{"method":"GET","path":"/api/tasks","status":200,"duration_ms":12,"ip":"192.168.1.1","user_agent":"Mozilla/5.0..."}
```

**Why:** Structured logs are required for log aggregation services (ELK, CloudWatch, Datadog) used in production at scale. Grep-friendly, machine-parseable, and filterable.

### 8. Health Check with DB Ping

**File:** `backend/internal/router/router.go`

`GET /api/health` now pings the database and returns pool statistics:

```json
{
  "status": "ok",
  "db_pool": {
    "total_conns": 100,
    "idle_conns": 85,
    "acquired_conns": 15
  }
}
```

Returns `503 Service Unavailable` if the database is unreachable.

**Why:** Load balancers (Nginx, ALB, K8s) need a reliable health endpoint to route traffic away from unhealthy instances. Pool stats enable real-time capacity monitoring.

### 9. Configurable CORS

**File:** `backend/internal/middleware/cors.go`

CORS now supports a whitelist of allowed origins via `ALLOWED_ORIGINS` environment variable:

```env
ALLOWED_ORIGINS=https://app.example.com,https://staging.example.com
```

**Why:** Wildcard CORS (`*`) is a security risk in production. At scale, you deploy the frontend to specific domains and only those should be allowed.

### 10. Production Server Timeouts

**File:** `backend/cmd/server/main.go`

| Timeout         | Default   | Purpose                              |
|-----------------|-----------|--------------------------------------|
| `READ_TIMEOUT`  | `15s`     | Maximum time to read request         |
| `WRITE_TIMEOUT` | `15s`     | Maximum time to write response       |
| `IDLE_TIMEOUT`  | `120s`    | Keep-alive timeout for idle conns    |
| `SHUTDOWN_TIMEOUT`| `30s`   | Graceful shutdown window             |

**Why:** Default Go HTTP server has no timeouts, meaning a slow client can hold a connection open indefinitely. At 50k users, this exhausts file descriptors.

---

## Getting Started

### Prerequisites

- [Go](https://go.dev/) ≥ 1.26
- [Node.js](https://nodejs.org/) ≥ 20
- [PostgreSQL](https://www.postgresql.org/) ≥ 15

### 1. Database Setup

```sql
CREATE DATABASE todo;
```

Run migrations in order:

```bash
psql -d todo -f backend/migrations/1_create_users.up.sql
psql -d todo -f backend/migrations/2_create_todos.up.sql
psql -d todo -f backend/migrations/3_add_indexes.up.sql
```

### 2. Backend Setup

```bash
cd backend

# Copy and configure environment
cp .env.example .env   # or edit .env directly

# Run the server
go run ./cmd/server
```

### 3. Frontend Setup

```bash
cd frontend
npm install
npm run dev
```

The Vite dev server proxies `/api` requests to `http://localhost:8000`.

---

## Environment Variables

All configuration is done through environment variables (or a `.env` file):

```env
# Required
DATABASE_URL=postgres://user:pass@localhost:5432/todo?sslmode=disable
JWT_SECRET=your-production-secret-here

# Optional (with defaults shown)
PORT=8000

# Database Pool
DB_MAX_CONNS=100
DB_MIN_CONNS=10
DB_MAX_CONN_LIFETIME=30m
DB_MAX_CONN_IDLE_TIME=5m
DB_HEALTH_CHECK_PERIOD=30s

# Rate Limiting
RATE_LIMIT_RPS=100
RATE_LIMIT_BURST=200

# Server Timeouts
READ_TIMEOUT=15s
WRITE_TIMEOUT=15s
IDLE_TIMEOUT=120s
SHUTDOWN_TIMEOUT=30s

# Security
MAX_REQUEST_BODY_SIZE=1048576
ALLOWED_ORIGINS=https://your-domain.com
```

---

## API Endpoints

| Method   | Endpoint               | Auth | Description                        |
|----------|------------------------|------|------------------------------------|
| `GET`    | `/api/health`          | No   | Health check + DB pool stats       |
| `POST`   | `/api/auth/register`   | No   | Create account                     |
| `POST`   | `/api/auth/login`      | No   | Login, get JWT                     |
| `GET`    | `/api/auth/me`         | Yes  | Current user info                  |
| `GET`    | `/api/tasks?limit=50&offset=0` | Yes | List tasks (paginated)    |
| `GET`    | `/api/tasks/{id}`      | Yes  | Get single task                    |
| `POST`   | `/api/tasks`           | Yes  | Create task                        |
| `PATCH`  | `/api/tasks/{id}`      | Yes  | Update task                        |
| `DELETE` | `/api/tasks/{id}`      | Yes  | Delete task                        |

---

## Production Deployment Checklist

- [ ] Set a strong, unique `JWT_SECRET` (≥32 characters)
- [ ] Set `ALLOWED_ORIGINS` to your frontend domain(s)
- [ ] Run migration `3_add_indexes.up.sql` to create performance indexes
- [ ] Configure `DB_MAX_CONNS` based on your PostgreSQL `max_connections`
- [ ] Put the server behind a reverse proxy (Nginx, Caddy, or a cloud LB)
- [ ] Enable TLS termination at the proxy layer
- [ ] Set up log aggregation (ELK, CloudWatch, etc.) for the structured JSON logs
- [ ] Monitor `/api/health` from your load balancer for active health checks
- [ ] Tune PostgreSQL `shared_buffers`, `work_mem`, and `effective_cache_size`
- [ ] Set OS-level file descriptor limits (`ulimit -n 65535`)
- [ ] Consider read replicas if read load exceeds a single PostgreSQL instance

---

## Project Structure

```
Todo-App/
├── backend/
│   ├── cmd/server/
│   │   └── main.go                    # Entry point
│   ├── internal/
│   │   ├── config/config.go           # Environment configuration
│   │   ├── database/database.go       # Pool-tuned DB connection
│   │   ├── handler/
│   │   │   ├── auth_handler.go        # Register, login, me
│   │   │   ├── task_handler.go        # CRUD + pagination
│   │   │   └── helpers.go            # JSON response helpers
│   │   ├── middleware/
│   │   │   ├── auth.go               # JWT authentication
│   │   │   ├── cors.go               # Configurable CORS
│   │   │   ├── limits.go             # Body size + request timeout
│   │   │   ├── logger.go             # Structured JSON logging
│   │   │   └── ratelimit.go          # Token-bucket rate limiter
│   │   ├── models/
│   │   │   ├── task.go               # Task + PaginatedResponse
│   │   │   └── user.go               # User + auth DTOs
│   │   ├── repository/
│   │   │   ├── task_repository.go    # Paginated DB queries
│   │   │   └── user_repository.go    # User DB queries
│   │   └── router/router.go          # Route + middleware wiring
│   ├── migrations/
│   │   ├── 1_create_users.up.sql
│   │   ├── 2_create_todos.up.sql
│   │   └── 3_add_indexes.up.sql      # Performance indexes
│   ├── .env
│   ├── go.mod
│   └── go.sum
├── frontend/
│   ├── src/
│   │   ├── api/index.ts              # Paginated API client
│   │   ├── context/AuthContext.tsx
│   │   ├── pages/
│   │   │   ├── DashboardPage.tsx     # Load More pagination
│   │   │   ├── LoginPage.tsx
│   │   │   └── RegisterPage.tsx
│   │   ├── components/
│   │   ├── types/index.ts            # PaginatedResponse type
│   │   └── App.tsx
│   ├── package.json
│   └── vite.config.ts
└── README.md
```

---

## License

MIT
