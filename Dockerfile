# Stage 1: Build the Go binary
FROM golang:1.23.4 AS builder
WORKDIR /app

# Copy go.mod and go.sum files
COPY go.mod go.sum ./

# Download all dependencies
RUN go mod download

# Copy the source code and public assets
COPY . .

# Build the Go app
RUN go build -o my-todo-app main.go

# Stage 2: Create a minimal image with just the binary
FROM gcr.io/distroless/base-debian10
WORKDIR /app

# Copy the Go binary and public assets from the builder stage
COPY --from=builder /app/my-todo-app .
COPY --from=builder /app/public /app/public
COPY --from=builder /app/.env .

# Expose the port the app listens on
EXPOSE 8080

# Set the entry point to the binary
ENTRYPOINT ["./my-todo-app"]
