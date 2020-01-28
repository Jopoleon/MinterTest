FROM golang:alpine as gobuilder

LABEL maintainer="Egor Miloserdov <egortictac@mail.ru>"

RUN apk update && apk add --no-cache git

WORKDIR /app

COPY go.mod go.sum ./

RUN go mod download

COPY . .

RUN GOOS=linux go build -o main .

FROM alpine:latest
RUN apk --no-cache add ca-certificates

WORKDIR /root/

COPY --from=gobuilder /app/main .
COPY --from=gobuilder /app/.env .

EXPOSE 8080

CMD ["./main"]