FROM golang:1.16-alpine

RUN apk add binutils

WORKDIR /app

COPY ./go.mod ./
COPY ./go.sum ./
RUN go mod download

COPY ./*.go ./

RUN go build -o /webhooks-sqs-worker

EXPOSE 9000

RUN chmod +x /webhooks-sqs-worker
CMD ["/webhooks-sqs-worker"]