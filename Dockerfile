# Dockerfile for the URL Service (main API)
FROM golang:1.26-alpine AS builder

WORKDIR /app
COPY go.mod ./
RUN go mod download || true

COPY . .
RUN go mod tidy && go build -o golink ./main.go

# Final image
FROM alpine:latest
WORKDIR /app
COPY --from=builder /app/golink .
EXPOSE 8080
CMD ["./golink"]
