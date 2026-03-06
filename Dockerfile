FROM golang:1-alpine AS builder

ENV CGO_ENABLED=0
ENV GOOS=linux
ENV GOARCH=amd64

WORKDIR /build

COPY go.mod go.sum ./
RUN go mod download

COPY model ./model/
COPY scraper ./scraper/
COPY utils ./utils/
COPY web ./web/
COPY *.go .

RUN go build -o main ./web/main.go

FROM alpine:latest AS release

# Install runtime dependencies (certificates, timezone, init)
RUN apk add --no-cache ca-certificates tzdata dumb-init

RUN addgroup -S appgroup && adduser -S appuser -G appgroup

COPY --from=builder /build/main /usr/local/bin/krip

EXPOSE 3000
USER appuser

HEALTHCHECK CMD wget --no-verbose --tries=1 --spider http://127.0.0.1:3000/_health || exit 1
ENTRYPOINT ["dumb-init", "--"]
CMD ["/usr/local/bin/krip"]
