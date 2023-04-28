FROM golang:1.20.3-alpine3.17 AS build_base

RUN apk add --no-cache git

WORKDIR /tmp/gowiki

COPY . .

RUN go build -o ./gowiki .


FROM alpine:3.17
RUN apk add ca-certificates

COPY --from=build_base /tmp/gowiki/* /app/

RUN chmod u+x /app/run.sh

CMD ["/app/run.sh"]