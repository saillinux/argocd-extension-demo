FROM golang:1.20.3-alpine3.17 AS build_base

ENV PATH $PATH:/root/google-cloud-sdk/bin

WORKDIR /tmp/extdemo

COPY . .

RUN go mod download

RUN go build -o ./extdemo .


FROM alpine:3.17
RUN apk add ca-certificates

COPY --from=build_base /tmp/extdemo/* /app/

RUN chmod u+x /app/run.sh

CMD ["/app/run.sh"]