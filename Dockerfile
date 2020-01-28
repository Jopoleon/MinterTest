FROM golang:alpine as builder

# ENV GO111MODULE=on

LABEL maintainer="Egor Miloserdov <egortictac@mail.ru>"

RUN apk update && apk add --no-cache git

# Set the current working directory inside the container
WORKDIR /app

# Copy go mod and sum files
COPY go.mod go.sum ./

# Download all dependencies. Dependencies will be cached if the go.mod and the go.sum files are not changed
RUN go mod download

# Copy the source from the current directory to the working Directory inside the container
COPY . .

# Build the Go app
#RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o main .
RUN GOOS=linux go build -o main .

# Start a new stage from scratch
FROM alpine:latest
RUN apk --no-cache add ca-certificates

WORKDIR /root/

COPY --from=builder /app/main .
COPY --from=builder /app/.env .

EXPOSE 8080

CMD ["./main"]