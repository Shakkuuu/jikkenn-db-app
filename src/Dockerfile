FROM golang:1.23

WORKDIR /app

COPY go.mod ./
COPY go.sum ./

RUN go mod download
RUN go mod tidy

COPY ./ ./

EXPOSE 8080

RUN go build -o /jikkenn-db

# 起動コマンド
CMD ["/jikkenn-db"]
