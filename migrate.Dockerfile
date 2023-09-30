FROM golang:1.19 AS builder

WORKDIR /notification-service/

COPY go.mod go.sum /notification-service/
RUN go mod download

COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -o ./migrate ./migrate/migrate.go

FROM alpine:latest

WORKDIR /root/

COPY --from=builder /notification-service/migrate /root/
COPY --from=builder /notification-service/configs /root/configs
COPY --from=builder /notification-service/app.docker.env /root/app.env

EXPOSE 8080

CMD [ "./migrate" ]
