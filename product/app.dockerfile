FROM golang:1.26.2-alpine AS builder

WORKDIR /src

COPY go.mod go.sum* ./
RUN go mod download

COPY product ./product

RUN CGO_ENABLED=0 GOOS=linux go build -o /product-service ./product/cmd/product

FROM alpine:3.22

RUN apk add --no-cache ca-certificates

COPY --from=builder /product-service /product-service

EXPOSE 8081

ENTRYPOINT ["/product-service"]
