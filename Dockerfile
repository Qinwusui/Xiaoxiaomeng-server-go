FROM golang:1.23-alpine as builder
COPY main.go /
COPY go.mod /
COPY go.sum /

WORKDIR /

RUN go build -ldflags="-s -w" -o /xxm .



FROM alpine:latest
COPY --from=builder /start ./
WORKDIR ./

EXPOSE 3456
RUN chmod 777 ./xxm
RUN ls
CMD ["./xxm"]