FROM node:22-alpine AS frontend

WORKDIR /app
COPY frontend/package.json frontend/package-lock.json ./
RUN npm ci --no-audit --no-fund
COPY frontend/ ./
RUN npm run build

FROM golang:1.25-bookworm AS backend

WORKDIR /src
COPY backend/go.mod backend/go.sum ./
RUN go mod download
COPY backend/ ./
RUN CGO_ENABLED=0 go build -o /out/server ./cmd/server

FROM debian:bookworm-slim

RUN apt-get update && apt-get install -y --no-install-recommends nginx ca-certificates openssl \
    && rm -rf /var/lib/apt/lists/* \
    && mkdir -p /etc/nginx/certs /var/www/html \
    && openssl req -x509 -nodes -newkey rsa:2048 -days 365 \
        -keyout /etc/nginx/certs/server.key \
        -out /etc/nginx/certs/server.crt \
        -subj "/CN=localhost"

COPY --from=backend /out/server /usr/local/bin/server
COPY --from=frontend /app/dist /var/www/html
COPY nginx.conf /etc/nginx/nginx.conf
COPY docker-entrypoint.sh /docker-entrypoint.sh
RUN chmod +x /docker-entrypoint.sh

ENV APP_DB_PATH=/data/app.db
ENV APP_HTTP_ADDR=127.0.0.1:8080

EXPOSE 8888

ENTRYPOINT ["/docker-entrypoint.sh"]
