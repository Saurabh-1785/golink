# Project Title

## GoLink — Scalable URL Shortener with Redis Caching and gRPC Analytics

---

# Short Project Description

GoLink is a distributed backend-focused URL shortening platform built using Go, Redis, PostgreSQL, and gRPC.

The system generates compact short URLs using Base62 encoding, optimizes redirect performance through Redis cache-aside caching, and uses gRPC-based internal service communication for asynchronous analytics processing.

The project is designed to demonstrate real-world backend engineering concepts such as:

* scalable service design
* caching strategies
* inter-service communication
* low-latency request handling
* modular backend architecture

---

# Main Features

## 1. URL Shortening

* Converts long URLs into compact short links
* Uses Base62 encoding for efficient short-code generation
* Ensures uniqueness and fast retrieval

Example:

```text id="cstc53"
https://example.com/very/long/url
↓
golink.io/aZ91Xk
```

---

# 2. Redis Cache-Aside Optimization

The Redirect Service first checks Redis before querying PostgreSQL.

Flow:

```text id="h98u3p"
Request → Redis Cache → DB Fallback → Cache Refill
```

Benefits:

* reduces database load
* improves redirect latency
* handles frequently accessed URLs efficiently

---

# 3. gRPC-Based Analytics Service

Instead of directly updating analytics in the main service, redirect events are sent to a dedicated Analytics Service using gRPC.

Analytics Service responsibilities:

* increment click count
* track timestamps
* maintain analytics data separately

This demonstrates:

* service-to-service communication
* modular backend design
* distributed system fundamentals

---

# 4. Dockerized Architecture

All services run inside Docker containers using Docker Compose.

Containers:

* URL Service
* Analytics Service
* Redis
* PostgreSQL

This simplifies:

* local development
* deployment
* environment consistency

---

# 5. REST API Endpoints

External communication uses REST APIs.

Main endpoints:

```http id="6edny5"
POST /shorten
GET /:shortCode
GET /analytics/:shortCode
```

---

# Project Architecture

```text id="52lphj"
                Client
                   ↓
            REST API Service
                   ↓
        ┌───────────────┐
        ↓               ↓
   Redis Cache      PostgreSQL
        ↓
   Redirect Handling
        ↓
   gRPC Analytics Service
        ↓
   Click Tracking Storage
```

---

# Technologies Used

| Component              | Technology  |
| ---------------------- | ----------- |
| Backend Language       | Go          |
| Database               | PostgreSQL  |
| Cache Layer            | Redis       |
| Internal Communication | gRPC        |
| Containerization       | Docker      |
| API Style              | REST + gRPC |

---

# Unique Aspects of the Project

Unlike traditional URL shortener projects, GoLink focuses heavily on backend scalability and distributed communication concepts.

Key differentiators:

* Redis cache-aside optimization
* gRPC-based internal microservice communication
* modular analytics processing
* Dockerized multi-service setup
* low-latency redirect architecture

Most student URL shorteners only implement CRUD functionality, while this project demonstrates production-oriented backend engineering patterns.

---

# Backend Concepts Demonstrated

This project showcases understanding of:

* distributed systems fundamentals
* service separation
* cache-aside pattern
* low-latency backend optimization
* REST vs gRPC communication
* containerized deployment
* database optimization
* scalable backend architecture

---

# Why gRPC Was Used

gRPC was chosen for internal service communication because:

* it uses HTTP/2
* supports efficient Protocol Buffers serialization
* provides lower latency than REST for internal services
* enables strongly typed contracts between services

This makes it ideal for backend microservice communication.

---

# Why Redis Was Used

Redis was used to reduce repeated database queries during URL redirection.

Benefits:

* faster redirects
* reduced PostgreSQL load
* better scalability for hot URLs
* improved response times

---

# How Short URLs Are Generated

The project uses Base62 encoding:

```text id="v7t9jk"
0-9 + a-z + A-Z
```

This generates compact and human-readable short codes while supporting a very large number of unique URLs.

---

# Resume Description (Short Version)

```text id="sjk0ep"
Built a scalable URL shortener backend in Go using Redis caching and gRPC-based analytics services, implementing cache-aside optimization, Base62 URL generation, and modular service communication using Dockerized architecture.
```

---

# Resume Description (Detailed Version)

```text id="klpxq7"
Developed a distributed URL shortening platform using Go, PostgreSQL, Redis, and gRPC. Implemented Redis cache-aside optimization for low-latency redirects, Base62-based short-code generation, and gRPC-powered analytics microservices for modular backend processing. Containerized the complete multi-service architecture using Docker Compose.
```

---

# Suggested Folder Structure

```text id="10xg5l"
golink/
│
├── api/
├── analytics-service/
├── proto/
├── redis/
├── postgres/
├── docker-compose.yml
├── go.mod
└── README.md
```

---

# Interview Talking Points

You should be able to explain:

* Why caching improves scalability
* Difference between REST and gRPC
* Cache-aside pattern
* Why Redis is faster than PostgreSQL for redirects
* Why service separation improves maintainability
* Base62 encoding logic
* Why Docker Compose was used

---

# Project Difficulty & Resume Impact

| Category            | Rating |
| ------------------- | ------ |
| Difficulty          | 7/10   |
| Resume Value        | 8.5/10 |
| Uniqueness          | 8/10   |
| Backend Learning    | 9/10   |
| Interview Potential | 8.5/10 |
