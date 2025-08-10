# --- Frontend builder ---
FROM oven/bun:1.2.19-alpine AS frontend-builder
WORKDIR /app

COPY frontend/package.json frontend/bun.lock ./
RUN bun install

COPY frontend/ .
RUN chmod +x build-all.sh
RUN bun run build

# --- Backend builder ---
FROM golang:1.24.6-alpine AS backend-builder
WORKDIR /app

COPY backend/go.mod backend/go.sum ./
RUN go mod download

COPY backend/ .
RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-s -w" -o backend-binary ./...

# --- Final stage ---
FROM scratch

COPY --from=backend-builder /app/backend-binary /backend-binary
COPY --from=frontend-builder /app/dist /public

CMD ["/backend-binary"]
