FROM oven/bun:1.4.3-alpine AS frontend-builder
WORKDIR /app/web
COPY web/package.json web/bun.lock ./
RUN bun install --frozen-lockfile
COPY web/ ./
RUN bun run build

FROM golang:1.26-alpine AS backend-builder
WORKDIR /app
COPY go.mod go.sum ./
RUN apk add --no-cache gcc musl-dev libc-dev && go mod download
COPY . .
# The frontend bundle is embedded into the binary, so it must be built first.
COPY --from=frontend-builder /app/web/dist ./web/dist
RUN CGO_ENABLED=1 go build -o tairitsu ./cmd/tairitsu

FROM alpine:3.22 AS production
RUN apk add --no-cache ca-certificates
COPY --from=backend-builder /app/tairitsu /app/tairitsu
WORKDIR /app
# The console listens on a fixed port. Map it wherever you like, e.g. -p 8080:3000.
EXPOSE 3000
CMD ["/app/tairitsu"]
