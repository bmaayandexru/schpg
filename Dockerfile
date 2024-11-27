# Dokerfile для компактной сборки (без возможности запуска тестов)
# Образ 35.83 Mb
# специальный образ для сборки
FROM golang:1.22.1-alpine as builder
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN go build -o exefile .
# начало сборки итогового образа. 
# используем Alpine для выполнения // просто пустой линукс??? # хотя лучше не latest
FROM alpine:latest
# поменяем для разнообразия рабочий каталог
WORKDIR /apps
# устанавливаем необходимые зависимости
RUN apk --no-cache add ca-certificates libc6-compat
# копируем скомпилированный бинарный файл и .env из стадии сборки
COPY --from=builder /app/web ./web
COPY --from=builder /app/scheduler.db .
COPY --from=builder /app/exefile .
# устанавливаем разрешения на выполнение бинарного файла
RUN chmod +x ./exefile

# порт справочно
EXPOSE 7540
# окружение
ENV TODO_PORT="7540"
ENV TODO_CONNSTR="scheduler.db"
ENV TODO_PASSWORD="12111"

# запуск бинарника
CMD ["./exefile"]
