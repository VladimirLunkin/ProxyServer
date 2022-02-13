FROM golang:1.16-alpine

WORKDIR /app

COPY go.mod ./
COPY go.sum ./
RUN go mod download

COPY . ./

RUN go build -o /docker-proxy

EXPOSE 8080

ENTRYPOINT [ "/docker-proxy" ]
