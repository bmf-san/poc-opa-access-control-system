FROM golang:1.24-alpine

WORKDIR /app

# Install build dependencies
RUN apk add --no-cache git

# Copy go mod files and internal module
COPY internal/ ./internal/
COPY cmd/pdp/go.mod cmd/pdp/go.sum ./

# Copy the source code
COPY cmd/pdp/ ./cmd/pdp/

# Set GOPROXY to handle network issues better
ENV GOPROXY=https://proxy.golang.org,direct

# Set the working directory to where the main package is
WORKDIR /app/cmd/pdp

# Create policy directory
RUN mkdir -p policy

# Copy policy files
COPY cmd/pdp/policy/*.rego policy/

# Build the application
RUN go mod download && \
    CGO_ENABLED=0 GOOS=linux go build -o /pdp .

# Use scratch as minimal base image
FROM scratch
COPY --from=0 /pdp /pdp
COPY --from=0 /app/cmd/pdp/policy /policy
EXPOSE 8081

ENTRYPOINT ["/pdp"]
