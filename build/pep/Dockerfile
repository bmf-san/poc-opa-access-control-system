FROM golang:1.24-alpine

WORKDIR /app

# Install build dependencies
RUN apk add --no-cache git

# Copy go mod files and internal module
COPY internal/ ./internal/
COPY cmd/pep/go.mod cmd/pep/go.sum ./

# Copy the source code
COPY cmd/pep/ ./cmd/pep/

# Set GOPROXY to handle network issues better
ENV GOPROXY=https://proxy.golang.org,direct

# Set the working directory to where the main package is
WORKDIR /app/cmd/pep

# Build the application
RUN go mod download && \
    CGO_ENABLED=0 GOOS=linux go build -o /pep .

# Use scratch as minimal base image
FROM scratch
COPY --from=0 /pep /pep
EXPOSE 80

ENTRYPOINT ["/pep"]
