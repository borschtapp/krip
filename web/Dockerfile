FROM golang:1-alpine AS builder

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN go build -o krip-web ./web/main.go

FROM alpine:latest

WORKDIR /app
COPY --from=builder /app/krip-web .

EXPOSE 3000
HEALTHCHECK CMD wget --no-verbose --tries=1 --spider http://localhost:3000/health || exit 1
ENTRYPOINT ["./krip-web"]
