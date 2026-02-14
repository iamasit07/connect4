# Stage 1: Build Frontend
FROM node:20-alpine AS frontend_builder
WORKDIR /app/frontend
COPY frontend/package.json frontend/package-lock.json ./
RUN npm install
COPY frontend/ .

# Build environment variables (Empty for relative paths in Monolith)
ARG VITE_BACKEND_URL=""
ARG VITE_WS_URL=""
ENV VITE_BACKEND_URL=$VITE_BACKEND_URL
ENV VITE_WS_URL=$VITE_WS_URL
RUN npm run build

# Stage 2: Build Backend
FROM golang:1.24-alpine AS backend_builder
WORKDIR /app/backend
# Install build dependencies
RUN apk add --no-cache git
COPY backend/go.mod backend/go.sum ./
RUN go mod download
COPY backend/ .
# Build the binary
RUN CGO_ENABLED=0 GOOS=linux go build -o main ./cmd/api

# Stage 3: Final Image
FROM alpine:latest
WORKDIR /root/
# Install ca-certificates for external API calls (e.g. Google OAuth)
RUN apk --no-cache add ca-certificates

# Copy Binary
COPY --from=backend_builder /app/backend/main .
# Copy Frontend Build to static directory
COPY --from=frontend_builder /app/frontend/dist ./static
# Copy .env example if needed (usually handled by Render env vars)

EXPOSE 8080
CMD ["./main"]
