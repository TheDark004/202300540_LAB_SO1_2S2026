# Usa imagen de Go para compilar dandole el tag de builder
FROM golang:1.25-alpine AS builder
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN go mod tidy
RUN CGO_ENABLED=0 GOOS=linux go build -o api3 main.go

# Usa imagen ligera para ejecutar
FROM alpine:3.22
WORKDIR /app
COPY --from=builder /app/api3 ./
EXPOSE 8083
CMD ["./api3"]