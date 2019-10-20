FROM golang:1.12-alpine

COPY . /src
WORKDIR /src

RUN apk add --no-cache --virtual .deps git build-base \
    && go build \
    && mv subrun / \
    && rm -rf /src \
    && apk del .deps

CMD ["/subrun"]
