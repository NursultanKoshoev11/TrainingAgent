FROM golang:1.22-alpine AS builder
ARG SERVICE
WORKDIR /src
COPY go.mod ./
COPY internal ./internal
COPY cmd ./cmd
RUN CGO_ENABLED=0 GOOS=linux go build -o /out/app ./cmd/${SERVICE}

FROM alpine:3.20
RUN adduser -D -H appuser
USER appuser
WORKDIR /app
COPY --from=builder /out/app /app/app
EXPOSE 8080 8081 8082 8083
ENTRYPOINT ["/app/app"]
