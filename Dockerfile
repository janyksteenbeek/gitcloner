FROM golang:1.23-alpine AS builder

WORKDIR /app
COPY . .

# Build with proper flags for better compatibility
ARG TARGETOS
ARG TARGETARCH
RUN CGO_ENABLED=0 GOOS=${TARGETOS} GOARCH=${TARGETARCH} \
    go build -a -installsuffix cgo -o gitcloner ./cmd/main

FROM alpine:latest

WORKDIR /app
COPY --from=builder /app/gitcloner .

EXPOSE 8080
CMD ["./gitcloner"] 