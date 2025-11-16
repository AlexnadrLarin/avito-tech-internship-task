FROM alpine:latest

RUN apk add --no-cache apache2-utils curl make

WORKDIR /app

COPY ab-tests/Makefile ./Makefile

CMD ["sh", "-c", "make init && make prepare && make ab"]

