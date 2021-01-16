FROM golang:alpine as builder

RUN mkdir uploads

RUN mkdir app
WORKDIR /app

# Copy all the sources file to working directory
COPY . .

# Building app into main executable
RUN go build -o main .

EXPOSE 8181

CMD ["./main"]