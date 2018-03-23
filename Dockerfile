FROM golang:1.10.0-alpine AS build

RUN apk add --no-cache gcc musl-dev make

WORKDIR /go/src/github.com/fukt/dweller

COPY . .

RUN make build

FROM alpine:3.7 AS run

RUN apk add --no-cache ca-certificates

WORKDIR /

COPY --from=build /go/src/github.com/fukt/dweller/bin/dweller /dweller

ENTRYPOINT /dweller
