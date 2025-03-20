FROM golang:1.24-alpine

WORKDIR /app

# Install build dependencies
RUN apk add --no-cache git

# Copy go mod files and internal module
COPY internal/ ./internal/
COPY cmd/pip/go.mod cmd/pip/go.sum ./

# Copy the source code
COPY cmd/pip/ ./cmd/pip/

# Set GOPROXY to handle network issues better
ENV GOPROXY=https://proxy.golang.org,direct

# Set the working directory to where the main package is
WORKDIR /app/cmd/pip

# Build the application
RUN go mod download && \
    CGO_ENABLED=0 GOOS=linux go build -o /pip .

# Use scratch as minimal base image
FROM scratch
COPY --from=0 /pip /pip
EXPOSE 8082

ENTRYPOINT ["/pip"]
