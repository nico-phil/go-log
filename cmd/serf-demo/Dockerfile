# Use the official Golang base image with a specific version and distribution
FROM golang:1.24-alpine AS builder
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-s -w" -o serf-node main.go

# Final image: distroless (no shell, no package manager, very small attack surface)
FROM gcr.io/distroless/base-debian12
WORKDIR /
COPY --from=builder /app/serf-node .
USER nonroot:nonroot
ENTRYPOINT ["/serf-node"]