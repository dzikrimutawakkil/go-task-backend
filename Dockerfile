# 1. Build Stage (Compiling the code)
FROM golang:1.24-alpine AS builder

WORKDIR /app

# Copy dependency files first (better caching)
COPY go.mod go.sum ./
RUN go mod download

# Copy the rest of the source code
COPY . .

# Build the application binary named "main"
RUN go build -o main .

# 2. Run Stage (Running the binary in a small container)
FROM alpine:latest

WORKDIR /root/

# Copy the binary from the builder stage
COPY --from=builder /app/main .

# Copy the .env file (Optional: usually env vars are set in docker-compose)
COPY --from=builder /app/.env . 

# Expose the port your app runs on
EXPOSE 8080

# Command to run the app
CMD ["./main"]