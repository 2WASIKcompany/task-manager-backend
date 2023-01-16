FROM golang:alpine AS builder
COPY . /build/

WORKDIR /build
ARG CGO_ENABLED=0
ARG GOOS=linux

RUN go build -installsuffix 'static' -o app cmd/main.go

FROM alpine:latest
COPY --from=builder /build/app .
EXPOSE 8080 8080
ENTRYPOINT ["./app"]