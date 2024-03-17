FROM golang:1.20-alpine as builder
WORKDIR /build
RUN apk add ca-certificates tzdata
COPY ./ ./
RUN go build ./cmd/coffee-app-bot

FROM alpine:3.17
WORKDIR /app
COPY --from=builder /build/coffee-app-bot ./coffee-app-bot
ENTRYPOINT ./coffee-app-bot