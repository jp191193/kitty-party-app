# Stage 1: Build the application
FROM golang:1.22-alpine AS builder

# Set the working directory inside the container
WORKDIR /app

# Install necessary build tools (optional but good for Alpine)
RUN apk add --no-cache git

# Copy go.mod and go.sum first to leverage Docker cache
COPY go.mod go.sum ./
RUN go mod download

# Copy the rest of the application code
COPY . .

# Build the Go application
# CGO_ENABLED=0 ensures a static binary
# -trimpath and -ldflags="-s -w" shrink the binary size and remove debug info
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -trimpath -ldflags="-s -w" -o /api ./cmd/api

# Stage 2: Create a minimal runner container
FROM alpine:latest

# Set timezone data and CA certificates (needed for external HTTPS requests if any)
RUN apk --no-cache add ca-certificates tzdata

WORKDIR /app

# Copy only the compiled binary from the builder stage
COPY --from=builder /api /app/api

# Tell Docker the container listens on port 8080
EXPOSE 8080

# Command to run the executable
CMD ["/app/api"]
