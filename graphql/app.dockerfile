FROM golang:1.26.2-alpine AS builder

WORKDIR /src

COPY go.mod go.sum* ./
RUN go mod download

COPY account ./account
COPY order ./order
COPY product ./product
COPY graphql ./graphql

RUN CGO_ENABLED=0 GOOS=linux go build -o /graphql-service ./graphql

FROM alpine:3.22

RUN apk add --no-cache ca-certificates

COPY --from=builder /graphql-service /graphql-service

EXPOSE 8080

ENTRYPOINT ["/graphql-service"]