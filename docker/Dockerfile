# docker/Dockerfile
FROM golang:1.23-alpine AS builder

WORKDIR /app

# Copy go mod file
COPY go.mod ./
# Copy go.sum if it exists
COPY go.sum ./

# Download dependencies
RUN go mod download

# Copy source code
COPY . .

# Build the application
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o main cmd/api/main.go

# Final stage
FROM alpine:latest

RUN apk --no-cache add ca-certificates tzdata

WORKDIR /root/


# Copy the binary from builder
COPY --from=builder /app/main .
COPY --from=builder /app/.env . 

# Expose port
EXPOSE 8080

CMD ["sh", "-c", "until pg_isready -h $DB_HOST -p $DB_PORT; do echo 'Waiting for postgres...'; sleep 2; done; echo 'PostgreSQL is ready'; ./main"]
