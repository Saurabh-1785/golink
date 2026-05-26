# GoLink — Scalable URL Shortener

A distributed URL shortening backend built with **Go**, **PostgreSQL**, **Redis**, and **gRPC**.

## Features

- **URL Shortening** — Converts long URLs into compact short links using Base62 encoding
- **Redis Cache-Aside** — Checks Redis before hitting PostgreSQL for fast redirects
- **gRPC Analytics** — Click tracking handled by a separate gRPC microservice
- **Dockerized** — All services run via Docker Compose

## Architecture

```
Client → REST API (port 8080)
              ├── Redis Cache (check first)
              ├── PostgreSQL (fallback + storage)
              └── gRPC → Analytics Service (port 50051)
                              └── PostgreSQL (click tracking)
```

## Tech Stack

| Component   | Technology |
|-------------|-----------|
| Language    | Go        |
| Database    | PostgreSQL|
| Cache       | Redis     |
| Internal RPC| gRPC      |
| Container   | Docker    |

## Quick Start

```bash
# Start all services
docker compose up --build

# Shorten a URL
curl -X POST http://localhost:8080/shorten \
  -H "Content-Type: application/json" \
  -d '{"url": "https://github.com"}'

# Redirect (open in browser or curl)
curl -L http://localhost:8080/<shortCode>

# View analytics
curl http://localhost:8080/analytics/<shortCode>
```

## API Endpoints

| Method | Endpoint               | Description              |
|--------|------------------------|--------------------------|
| POST   | `/shorten`             | Create a short URL       |
| GET    | `/{shortCode}`         | Redirect to original URL |
| GET    | `/analytics/{shortCode}` | Get click stats        |

## Project Structure

```
golink/
├── main.go                    # URL Service entry point
├── api/
│   ├── base62.go              # Base62 short code generator
│   ├── handlers.go            # HTTP handlers (shorten, redirect, analytics)
│   └── router.go              # Route definitions
├── analytics-service/
│   └── main.go                # gRPC Analytics Service
├── proto/
│   ├── analytics.proto        # gRPC service definition
│   ├── analytics.pb.go        # Message types
│   ├── analytics_grpc.pb.go   # gRPC client/server stubs
│   └── codec.go               # JSON codec for gRPC
├── postgres/
│   ├── schema.sql             # Database schema
│   └── postgres.go            # DB connection & queries
├── redis/
│   └── redis.go               # Redis cache helpers
├── Dockerfile                 # URL Service container
├── Dockerfile.analytics       # Analytics Service container
├── docker-compose.yml         # Orchestration
└── README.md
```

## How It Works

### URL Shortening
1. Client sends a long URL via `POST /shorten`
2. A random 6-char Base62 code is generated
3. Mapping is saved to PostgreSQL and cached in Redis
4. Short URL is returned

### Redirect (Cache-Aside Pattern)
1. Client visits `GET /{shortCode}`
2. **Step 1**: Check Redis cache → if found, redirect immediately (cache HIT)
3. **Step 2**: If not in cache, query PostgreSQL (cache MISS)
4. **Step 3**: Store result in Redis for future requests (cache refill)
5. **Step 4**: Record click via gRPC to Analytics Service

### Analytics
- Each redirect triggers a gRPC call to the Analytics Service
- Analytics Service increments click count in PostgreSQL
- Stats are available via `GET /analytics/{shortCode}`

## Environment Variables

| Variable             | Default          | Description                  |
|----------------------|------------------|------------------------------|
| PORT                 | 8080             | HTTP server port             |
| DB_HOST              | localhost        | PostgreSQL host              |
| DB_PORT              | 5432             | PostgreSQL port              |
| DB_USER              | golink           | PostgreSQL user              |
| DB_PASSWORD          | golink123        | PostgreSQL password          |
| DB_NAME              | golink           | PostgreSQL database name     |
| REDIS_ADDR           | localhost:6379   | Redis address                |
| ANALYTICS_GRPC_ADDR  | localhost:50051  | Analytics gRPC service addr  |
| GRPC_PORT            | 50051            | Analytics gRPC server port   |

## Stop Services

```bash
docker compose down
docker compose down -v  # also remove database volume
```
