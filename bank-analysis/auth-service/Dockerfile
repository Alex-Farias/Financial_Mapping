FROM golang:1.19-alpine as builder

WORKDIR /app

# Copy go.mod first (without requiring go.sum)
COPY go.mod ./
# Create empty go.sum if it doesn't exist
RUN touch go.sum

# Install dependencies
RUN go mod download

# Copy the source code
COPY . .

# Build the application
RUN go build -o auth-service .

# Use a smaller image for the final build
FROM alpine:latest

# Install CA certificates for HTTPS
RUN apk --no-cache add ca-certificates

WORKDIR /app

# Copy the binary from the builder stage
COPY --from=builder /app/auth-service .

# Expose the port
EXPOSE 8081

# Run the application
CMD ["./auth-service"]