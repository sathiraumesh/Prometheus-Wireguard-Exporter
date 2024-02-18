FROM golang:1.20 AS builder

WORKDIR /app

COPY . .

RUN go build -o main  -mod=vendor .


FROM ubuntu:23.04

COPY ./wrieguard_setup.sh .

RUN chmod +x ./wrieguard_setup.sh && \
    ./wrieguard_setup.sh

WORKDIR /opt/

COPY --from=builder /app/main .
