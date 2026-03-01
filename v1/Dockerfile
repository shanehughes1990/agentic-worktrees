FROM golang:1.25.3-alpine AS builder

WORKDIR /src

RUN apk add --no-cache git ca-certificates

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o /out/api ./cmd/api
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o /out/worker ./cmd/worker

FROM alpine:3.22 AS runtime

RUN apk add --no-cache ca-certificates git

WORKDIR /app

FROM runtime AS api

COPY --from=builder /out/api /app/api
EXPOSE 8080
ENTRYPOINT ["/app/api"]

FROM runtime AS worker

COPY --from=builder /out/worker /app/worker
EXPOSE 8081
ENTRYPOINT ["/app/worker"]