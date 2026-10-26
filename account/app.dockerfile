FROM golang:1.26.2-alpine AS builder

WORKDIR /src

COPY go.mod go.sum* ./
RUN go mod download

COPY account ./account
COPY internal ./internal

RUN CGO_ENABLED=0 GOOS=linux go build -o /account-service ./account/cmd/account

FROM alpine:3.22

RUN apk add --no-cache ca-certificates

COPY --from=builder /account-service /account-service

EXPOSE 8080

ENTRYPOINT ["/account-service"]
