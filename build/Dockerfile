# syntax=docker/dockerfile:1
# docker build --rm -f build/Dockerfile -t source.toby3d.me/toby3d/hub .

# Build
FROM golang:alpine AS builder

WORKDIR /app

ENV CGO_ENABLED=0
ENV GOFLAGS="-mod=vendor -buildvcs=true"

COPY go.mod go.sum *.go ./
COPY internal ./internal/
COPY vendor ./vendor/
COPY web ./web/

RUN go build -o ./hub

# Run
FROM scratch

WORKDIR /

COPY --from=builder /app/hub /hub

EXPOSE 3000

ENTRYPOINT ["/hub"]
