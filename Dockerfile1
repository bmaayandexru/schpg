# Dokerfile для обычной сборки (с возможностью запуска тестов)
# Образ 775.15 Mb
FROM golang:1.22

WORKDIR /app

COPY go.mod go.sum ./

RUN go mod download

COPY . .

EXPOSE 7540

ENV TODO_PORT="7540"
ENV TODO_DBFILE="scheduler.db"
ENV TODO_PASSWORD="12111"

RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o exefile .

CMD ["./exefile"]
