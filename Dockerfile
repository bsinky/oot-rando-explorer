# STAGE 1: Download Tailwind CLI
FROM alpine:3.20 AS tailwind
WORKDIR /build
RUN apk add --no-cache curl
# Download the standalone binary
RUN curl -sLO https://github.com/tailwindlabs/tailwindcss/releases/latest/download/tailwindcss-linux-x64 \
    && chmod +x tailwindcss-linux-x64

# STAGE 2: Build the assets and binary
FROM golang:1.24-alpine AS builder
WORKDIR /app

# 1. Install gcompat so the Tailwind binary can run on Alpine/musl
RUN apk add --no-cache gcompat

# 2. Copy tailwind binary from previous stage
COPY --from=tailwind /build/tailwindcss-linux-x64 /usr/local/bin/tailwindcss

COPY go.mod go.sum ./
RUN go mod download
COPY . .

# 3. Run Tailwind compilation using the absolute path
RUN /usr/local/bin/tailwindcss -i ./assets/input.css -o ./assets/index.css --minify

# Build Go binary
RUN CGO_ENABLED=0 GOOS=linux go build -o sohrando .

# STAGE 3: Final lean image
FROM alpine:3.20
WORKDIR /root/
# We don't need gcompat here because the Go binary is statically linked (CGO_ENABLED=0)
COPY --from=builder /app/sohrando .
COPY --from=builder /app/assets ./assets
COPY --from=builder /app/templates ./templates

EXPOSE 8080
CMD ["./sohrando"]