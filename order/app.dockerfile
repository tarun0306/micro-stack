FROM golang:1.26.2-alpine AS builder

WORKDIR /src

COPY go.mod go.sum* ./
RUN go mod download

COPY account ./account
COPY order ./order
COPY product ./product

RUN CGO_ENABLED=0 GOOS=linux go build -o /order-service ./order/cmd/order

FROM alpine:3.22

RUN apk add --no-cache ca-certificates

COPY --from=builder /order-service /order-service

EXPOSE 8082

ENTRYPOINT ["/order-service"]
