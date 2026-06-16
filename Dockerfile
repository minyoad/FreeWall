# Build stage
FROM golang:1.25-alpine AS builder

# Install build dependencies for CGO and Node.js for frontend build
RUN apk add --no-cache gcc musl-dev nodejs npm

WORKDIR /build

# Copy and download Go modules
COPY go.mod go.sum ./
RUN go mod download

# Copy source code
COPY ./ ./

# Build frontend assets so embed can include frontend/dist
RUN cd frontend && npm ci && npm run build

# Build the binary
ARG TARGETARCH
RUN CGO_ENABLED=1 GOOS=linux GOARCH=${TARGETARCH:-amd64} CGO_CFLAGS="-D_LARGEFILE64_SOURCE" go build -tags with_utls,with_quic,with_clash_api,with_wireguard,with_gvisor -o /main .

# Final stage
FROM alpine:latest

WORKDIR /

COPY --from=builder /main /freegfw

EXPOSE 80
EXPOSE 443

CMD ["/freegfw"]
