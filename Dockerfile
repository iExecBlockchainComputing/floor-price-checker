# syntax=docker/dockerfile:1
FROM golang:1.16.0-alpine3.13

COPY ./src /app

RUN go build -buildmode=exe -o /app/nft-price-checker /app/app.go

ENTRYPOINT ["/app/nft-price-checker"]