# STAGE 1: Build the binary
FROM golang:1.24-alpine AS builder
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -o sohrando .

# STAGE 2: Final lean image
FROM alpine:3.20
WORKDIR /root/
# Copy only the binary from the builder
COPY --from=builder /app/sohrando .
# Copy static assets
COPY --from=builder /app/assets ./assets
COPY --from=builder /app/templates ./templates

EXPOSE 8080
CMD ["./sohrando"]
