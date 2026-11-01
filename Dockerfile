# One Dockerfile, two images. Pick one with a build target:
#   docker build --target analytics-service -t golink-analytics .
#   docker build --target url-service       -t golink-url .
# docker-compose.yml sets `target:` for each service.

# ---- Shared build stage: download deps once, compile both binaries ----
FROM golang:1.26-alpine AS builder

WORKDIR /app
# Copy module files first so the dependency layer is cached
# and only re-downloaded when go.mod/go.sum change.
COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN go build -o /out/golink ./main.go \
 && go build -o /out/analytics-service ./analytics-service

# ---- Image 1: Analytics gRPC Service ----
FROM alpine:latest AS analytics-service
WORKDIR /app
COPY --from=builder /out/analytics-service .
EXPOSE 50051
CMD ["./analytics-service"]

# ---- Image 2: URL Service (main API) ----
# Kept last so a plain `docker build .` produces the URL service.
FROM alpine:latest AS url-service
WORKDIR /app
COPY --from=builder /out/golink .
EXPOSE 8080
CMD ["./golink"]
