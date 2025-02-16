FROM golang:1.23-alpine AS builder

WORKDIR /app
COPY . .
RUN go build -o gitcloner ./cmd/main

FROM alpine:latest

WORKDIR /app
COPY --from=builder /app/gitcloner .

EXPOSE 8080
CMD ["./gitcloner"] 