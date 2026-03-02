FROM golang:1.23-alpine AS builder

RUN apk add --no-cache gcc musl-dev

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN CGO_ENABLED=1 go build -o tripshare ./cmd/server

# ---

FROM alpine:3.20

RUN apk add --no-cache ca-certificates

WORKDIR /app

COPY --from=builder /app/tripshare .
COPY --from=builder /app/web ./web

RUN mkdir -p /app/data

ENV PORT=8080
ENV DB_PATH=/app/data/tripshare.db

EXPOSE 8080

CMD ["./tripshare"]
