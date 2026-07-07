# Stage 1: build backend
FROM golang:1.25-alpine AS backend
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -o server ./cmd/server

# Stage 2: final image
FROM alpine:3.21
RUN apk add --no-cache ca-certificates tzdata
WORKDIR /app
COPY --from=backend /app/server .
COPY frontend/dist ./frontend/dist
RUN mkdir -p /data
ENV DATA_FILE=/data/app.db
ENV ADDR=:8080
ENV TZ=Asia/Ho_Chi_Minh
EXPOSE 8080
CMD ["./server"]
