# syntax=docker/dockerfile:1.2
FROM golang:1.16.0-alpine3.13

WORKDIR /app

COPY go.mod ./
COPY go.sum ./

RUN go mod download

COPY ./src /app

RUN apk --update add gcc build-base gcc-go

RUN go build -compiler=gccgo -buildmode=exe -o /app/nft-price-checker /app/nft-price-checker.go

ENTRYPOINT ["/app/nft-price-checker"]