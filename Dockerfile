# Stage 1: Build the Go binary
FROM golang:1.23-alpine AS builder

# Set the working directory inside the container
WORKDIR /app

# Copy the Go module files and download dependencies
COPY go.mod go.sum ./
RUN go mod download

# Copy the rest of the source code
COPY . .

# Build the Enterprise Chain binary
# CGO_ENABLED=0 ensures a static binary that runs anywhere
RUN CGO_ENABLED=0 GOOS=linux go build -o enterprise-node .

# Stage 2: Create the minimal runtime image
FROM alpine:latest

WORKDIR /app

# Copy the compiled binary from the builder stage
COPY --from=builder /app/enterprise-node .

# Create a directory for the blockchain database
RUN mkdir -p /app/chaindata

# Expose the P2P and API ports
EXPOSE 3000
EXPOSE 8080

# The default command when the container starts
ENTRYPOINT ["./enterprise-node"]