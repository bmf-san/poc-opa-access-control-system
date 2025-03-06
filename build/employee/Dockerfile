FROM golang:1.24-alpine

WORKDIR /app

# Install build dependencies
RUN apk add --no-cache git

# Copy necessary Go files
COPY cmd/employee/go.mod cmd/employee/go.sum ./
COPY internal/ ./internal/
COPY cmd/employee/ ./cmd/employee/

# Set GOPROXY to handle network issues better
ENV GOPROXY=https://proxy.golang.org,direct

# Set the working directory to where the main package is
WORKDIR /app/cmd/employee

# Build the application
RUN go mod download && \
    CGO_ENABLED=0 GOOS=linux go build -o /employee .

# Use scratch as minimal base image
FROM scratch
COPY --from=0 /employee /employee
EXPOSE 8083

ENTRYPOINT ["/employee"]
