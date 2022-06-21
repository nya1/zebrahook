# syntax=docker/dockerfile:1
FROM golang:1.18-buster AS builder

WORKDIR /app

# Copy over all go config (go.mod, go.sum etc.)
COPY go.* ./

# Install any required modules
RUN go mod download
RUN go mod verify

# Copy over Go source code
COPY . ./

# Run the Go build and output binary
RUN CGO_ENABLED=0 go build -o /zebrahook zebrahook/cmd/zebrahook

FROM golang:1.18-alpine

WORKDIR /

COPY --from=builder /zebrahook /zebrahook

EXPOSE 3000

ENTRYPOINT ["/zebrahook"]
