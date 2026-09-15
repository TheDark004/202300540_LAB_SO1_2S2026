FROM golang:1.25-alpine AS builder
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN go mod tidy
RUN CGO_ENABLED=0 GOOS=linux go build -o api2 main.go

FROM alpine:3.22
WORKDIR /app
COPY --from=builder /app/api2 ./
EXPOSE 8082
CMD ["./api2"]