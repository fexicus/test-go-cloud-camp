
# Go HTTP Balancer

Этот проект реализует HTTP-балансировщик нагрузки с поддержкой round-robin, health-check и rate-limiting (Token Bucket).

В корне проекта есть Makefile с основными командами:
- make build - собрать бинарник балансировщика
 ```bash
   go build -o bin/pickpoint ./cmd/pickpoint/
```
- make run - запустить локально программу
 ```bash
   make build
   CONFIG_PATH=./config/local.yaml ./bin/pickpoint
```
- make docker-up - поднять всё через Docker Compose (DB, бэкенды, балансировщик)
 ```bash
    docker-compose -f docker-compose.yaml up
```
- make migrate-up / make migrate-down - применить (откатить) миграции в Postgres
 ```bash
    make migrate-up
    make migrate-down
```
## Запросы для проверки работоспособности программы
#### Первый запрос:
```http
  curl -v http://localhost:8080/
```
#### Второй запрос с повторением 10-ти раз:
```http
  for i in {1..10}; do curl -s http://localhost:8080/ && echo; done
```
#### Третий запрос с имуляцией разных клиентов:
```http
  curl -v -H "X-Forwarded-For: 127.0.0.1" http://localhost:8080/
```
## Конфигурация и логи
- Логи никуда не записываются кроме консоли, но можно доработать и писать их сразу в БД, чтобы сделать подобие аудита
- Все данных для БД и портов находятся в файле config.yaml
- Реализован Graceful Shutdown
- Также необходимо запустить другие бэки, чтобы можно было посмотреть работоспособность программы в другом терминале, перейдя в директорию **test_backs**. Команда - go run main.go -port=**номера порта (8081, 8082, 8083)**

