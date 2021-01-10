FROM golang:alpine as builder

# Installing git for downloading dependencies
RUN apk update && apk add --no-cache git

RUN mkdir app
WORKDIR /app

# Copy all the sources file to working directory
COPY . .

# Downloading dependencies (we already copied go.mod and go.sum)
RUN go mod download

# Building app into main executable
RUN go build -o main .

EXPOSE 8080

CMD ["./main"]
