FROM golang:1.9.3-alpine AS build

RUN apk add --no-cache gcc musl-dev make

WORKDIR /go/src/github.com/fukt/dweller

COPY . .

RUN make build

FROM alpine:3.7 AS run

WORKDIR /

COPY --from=build /go/src/github.com/fukt/dweller/bin/dweller /dweller

ENTRYPOINT /dweller
