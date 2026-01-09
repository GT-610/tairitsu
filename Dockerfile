FROM node:22-alpine AS frontend-builder
WORKDIR /app
COPY web/ ./
RUN npm ci && npm run build

FROM golang:1.25-alpine AS backend-builder
WORKDIR /app
COPY . .
RUN apk add --no-cache gcc musl-dev libc-dev && \
    go mod download && \
    CGO_ENABLED=1 go build -o tairitsu ./cmd/tairitsu

FROM nginx:alpine-slim AS production
COPY --from=frontend-builder /app/dist /usr/share/nginx/html
COPY --from=backend-builder /app/tairitsu /app/
COPY docker/nginx.conf /etc/nginx/conf.d/default.conf
WORKDIR /app
EXPOSE 3000
CMD ["sh", "-c", "nginx -g 'daemon off;' & /app/tairitsu"]
