FROM golang:1.22-alpine

COPY . /app

RUN apk update && \
    apk upgrade && \
    apk add php php-sockets && \
    adduser -D appuser && \
    chown -R appuser /app

WORKDIR /app

USER appuser

EXPOSE 8080
EXPOSE 9999
EXPOSE 7777

CMD ["go", "run", "."]