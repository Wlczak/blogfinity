FROM golang:1.24.5-alpine AS builder

WORKDIR /app 

COPY ./ /app

RUN go build -o blogfinity . && chmod +x blogfinity

FROM alpine:latest

WORKDIR /app

COPY --from=builder /app /app

CMD ["./blogfinity"]