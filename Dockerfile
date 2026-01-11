FROM golang:1.25-alpine AS backend-builder
WORKDIR /app
COPY backend/go.mod backend/go.sum ./
RUN go mod download
COPY backend/ .
RUN go build -o main ./cmd/server/main.go

FROM node:25-alpine AS frontend-builder
WORKDIR /app
COPY frontend/package*.json ./
RUN npm ci
COPY frontend/ .
RUN npm run build

FROM node:25-alpine
RUN apk add --no-cache caddy
WORKDIR /app
COPY --from=backend-builder /app/main ./backend-server
COPY --from=frontend-builder /app/build ./build
COPY --from=frontend-builder /app/package.json ./
COPY --from=frontend-builder /app/package-lock.json ./
RUN npm ci --omit=dev
COPY Caddyfile /etc/caddy/Caddyfile
EXPOSE 80

RUN echo "#!/bin/sh" > start.sh && \
    echo "./backend-server &" >> start.sh && \
    echo "PORT=3000 node build &" >> start.sh && \
    echo "caddy run --config /etc/caddy/Caddyfile --adapter caddyfile" >> start.sh && \
    chmod +x start.sh
CMD ["./start.sh"]