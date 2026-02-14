# Build Stage
FROM golang:1.24-alpine AS builder

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .

# Build the application
# -o main: output binary name
# ./cmd/api: entry point
RUN CGO_ENABLED=0 GOOS=linux go build -o main ./cmd/api

# Run Stage
FROM alpine:latest

WORKDIR /app

# Copy the pre-built binary file from the previous stage
COPY --from=builder /app/main .
COPY --from=builder /app/script ./script

# Expose port 8080 to the outside world
EXPOSE 8080

# Command to run the executable
CMD ["./main"]
