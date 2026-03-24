FROM golang:1.24-alpine AS build
WORKDIR /src

RUN apk add --no-cache git ca-certificates

COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o /out/server ./main.go

FROM golang:1.24-alpine AS dev
WORKDIR /app

RUN apk add --no-cache git

COPY go.mod go.sum ./
RUN go mod download

COPY . .

EXPOSE 8080
CMD ["go", "run", "main.go"]

FROM alpine:3.21 AS production
WORKDIR /app

RUN apk add --no-cache ca-certificates tzdata && \
    adduser -D -u 10001 appuser

COPY --from=build /out/server /app/server
COPY config.yaml /app/config.yaml

RUN mkdir -p /app/uploads /app/tmp && chown -R appuser:appuser /app

USER appuser
EXPOSE 8080
CMD ["/app/server"]
