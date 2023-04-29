FROM golang:1.20.3-alpine3.17 AS build_base

RUN apk add --no-cache git

WORKDIR /tmp/extdemo

COPY . .

RUN go build -o ./extdemo .


FROM alpine:3.17
RUN apk add ca-certificates

COPY --from=build_base /tmp/extdemo/* /app/

RUN chmod u+x /app/run.sh

CMD ["/app/run.sh"]