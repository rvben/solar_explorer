FROM golang:1.17.4
WORKDIR /app
COPY . .
RUN go get -d
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o app .

FROM alpine:latest
WORKDIR /app/
COPY --from=0 /app/app .
COPY config.yml.example /app/config.yml
CMD ["./app"]
