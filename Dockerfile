# Stage 1: build frontend
FROM node:22-alpine AS frontend
WORKDIR /app/frontend
COPY frontend/package.json ./
RUN npm install vite@5.4.19 @vitejs/plugin-vue@2.3.4 vue@3.5.17
COPY frontend/ ./
RUN npm run build

# Stage 2: build backend (modernc sqlite is pure Go, no CGO needed)
FROM golang:1.25-alpine AS backend
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -o server ./cmd/server

# Stage 3: final image
FROM alpine:3.21
RUN apk add --no-cache ca-certificates tzdata
WORKDIR /app
COPY --from=backend /app/server .
COPY --from=frontend /app/frontend/dist ./frontend/dist
RUN mkdir -p /data
ENV DATA_FILE=/data/app.db
ENV ADDR=:8080
EXPOSE 8080
CMD ["./server"]
