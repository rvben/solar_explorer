FROM golang:1.20.1 AS builder
WORKDIR /app
COPY . .
RUN go get -d
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o app .

FROM alpine:latest
WORKDIR /app/
COPY --from=builder /app/app .
COPY config.yml.example /app/config.yml
CMD ["./app"]
