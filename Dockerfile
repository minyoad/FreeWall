# Frontend build stage
FROM node:24-bullseye AS frontend-builder

WORKDIR /build/frontend
COPY frontend/package.json frontend/package-lock.json ./
RUN npm ci

COPY frontend/ ./
RUN npm run build

# Go build stage
FROM golang:1.25-alpine AS builder

# Install build dependencies for CGO
RUN apk add --no-cache gcc musl-dev

WORKDIR /build

# Copy and download Go modules
COPY go.mod go.sum ./
RUN go mod download

# Copy source code
COPY ./ ./

# Copy built frontend assets from frontend-builder
COPY --from=frontend-builder /build/frontend/dist ./frontend/dist

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
