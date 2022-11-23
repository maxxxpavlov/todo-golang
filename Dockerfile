FROM golang:1.16-alpine

WORKDIR /app

COPY go.mod ./
COPY go.sum ./

RUN go mod download

COPY *.go ./
COPY *.key ./
RUN go build -o /todo

CMD ["/todo"]
EXPOSE 3000