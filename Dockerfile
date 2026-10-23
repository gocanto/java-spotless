# syntax=docker/dockerfile:1

FROM golang:1.26.10-alpine AS builder
WORKDIR /src

COPY go.mod ./
COPY cmd ./cmd
COPY pkg ./pkg

RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -trimpath -ldflags="-s -w" -o /out/spotless ./cmd/spotless

FROM alpine:3.22
RUN apk add --no-cache ca-certificates git docker-cli docker-cli-compose
WORKDIR /work

COPY --from=builder /out/spotless /usr/local/bin/spotless

ENTRYPOINT ["spotless"]
CMD ["help"]
