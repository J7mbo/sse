# Accept the Go version for the image to be set as a build argument.
# Default to Go 1.10
ARG GO_VERSION=1.14

# First stage: build the executable.
FROM golang:${GO_VERSION}-alpine AS builder

# Install the Certificate-Authority certificates for the app to be able to make
# calls to HTTPS endpoints.
RUN apk add --no-cache ca-certificates git curl

ENV CGO_ENABLED=0

# Set the working directory outside $GOPATH to enable the support for modules.
WORKDIR /sse

COPY go.mod .
COPY go.sum .
RUN go mod download

# Import the code from the context.
COPY . /sse

# Build the executable to `/sse`.
RUN go build -o ./bin/sse .

# Final stage: the running container.
FROM golang:${GO_VERSION} AS final

WORKDIR /sse

# Import the Certificate-Authority certificates for enabling HTTPS.
COPY --from=builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/

# Import the compiled executable from the second stage.
COPY --from=builder /sse/bin/sse /sse

## Run the compiled binary.
ENTRYPOINT ["./sse"]