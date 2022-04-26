# syntax=docker/dockerfile:1
FROM golang:1.18.1
WORKDIR /src

COPY go.mod ./

RUN go mod download

COPY src/*.go ./

RUN go build -o /docker-gs-ping

EXPOSE 8080

CMD [ "/docker-gs-ping" ]