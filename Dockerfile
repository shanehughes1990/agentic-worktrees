# syntax=docker/dockerfile:1

FROM golang:1.25-alpine AS builder
WORKDIR /src

RUN apk add --no-cache git ca-certificates

COPY go.mod ./
RUN go mod download

COPY . .

ARG APP_PACKAGE=./cmd/agentic-worktrees
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -trimpath -ldflags="-s -w" -o /out/agentic-worktrees ${APP_PACKAGE}

FROM alpine:3.21
WORKDIR /app

RUN addgroup -S app && adduser -S app -G app && \
	apk add --no-cache ca-certificates git openssh-client

COPY --from=builder /out/agentic-worktrees /app/bin/agentic-worktrees

USER app
ENTRYPOINT ["/app/bin/agentic-worktrees"]
