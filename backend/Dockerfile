FROM golang:1.22-alpine AS builder

WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN CGO_ENABLED=0 go build -o sismoserver main.go scraper.go api.go db.go push.go middleware.go

FROM alpine:3.19

WORKDIR /app
RUN apk add --no-cache ca-certificates wget

COPY --from=builder /app/sismoserver .
RUN mkdir -p /app/data

ENV PORT=8082
EXPOSE 8082

VOLUME ["/app/data"]

ENTRYPOINT ["./sismoserver"]
