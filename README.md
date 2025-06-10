# Worker Pool



`internal/workerpool` — минимальная, но гибкая реализация динамического пула воркеров на Go. Для примера в проекте есть 
HTTP‑сервер (`main.go`) показывает, как пользоваться пулом в реальном приложении.

Основные фичи:

* **Автоскейлинг вверх/вниз** — создаём горутины, когда очередь растёт, и убираем при простое
* **Безопасное завершение** — ловим `SIGINT`/`SIGTERM`
* **Защита от panic** — падение задачи не роняет весь пул.
* **Быстро** — атомарные операции, счетчики без мьютекса.

---

## Запуск

```bash
$ go build -o server .

$ ./server -initial 5 -max 15 -buffer 100 -addr :8080
```

---

## Параметры демо‑сервера

| Флаг       | Значение по умолчанию | Описание                         |
| ---------- | --------------------- | -------------------------------- |
| `-initial` | `5`                   | Стартовое количество воркеров    |
| `-max`     | `10`                  | Верхний предел активных воркеров |
| `-buffer`  | `100`                 | Размер канала задач              |
| `-addr`    | `:8080`               | Адрес HTTP‑сервера               |

---


## Тестирование

Нагрузочное тестирование с siege
```bash
$ siege -c100 -t25S --content-type "application/json" 'http://localhost:8080 POST {"A": 1, "B": 2}'
```
### Пример:
#### siege log
```
** SIEGE 4.1.7
** Preparing 100 concurrent users for battle.
The server is now under siege...
HTTP/1.1 200     3.01 secs:      13 bytes ==> POST http://localhost:8080
HTTP/1.1 200     3.01 secs:      13 bytes ==> POST http://localhost:8080
...
HTTP/1.1 200     9.02 secs:      13 bytes ==> POST http://localhost:8080
HTTP/1.1 200     9.02 secs:      13 bytes ==> POST http://localhost:8080

TLifting the server siege...
Transactions:                 120    hits
Availability:                 100.00 %
Elapsed time:                  25.01 secs
Data transferred:               0.00 MB
Response time:               5627.50 ms
Transaction rate:               4.80 trans/sec
Throughput:                     0.00 MB/sec
Concurrency:                   27.00
Successful transactions:      120
Failed transactions:            0
Longest transaction:         6010.00 ms
Shortest transaction:        3010.00 ms
```
### server log
```bash
$ ./server -initial 5 -max 15 -buffer 100 -addr :8080
2025/06/10 22:06:39 Starting new worker with id 1
2025/06/10 22:06:39 Starting new worker with id 2
2025/06/10 22:06:39 Starting new worker with id 3
2025/06/10 22:06:39 Starting new worker with id 4
2025/06/10 22:06:39 Starting new worker with id 5
2025/06/10 22:06:39 Server listening on :8080
2025/06/10 22:06:41 Received task: 1 + 2
2025/06/10 22:06:41 Received task: 1 + 2
2025/06/10 22:06:41 Received task: 1 + 2
2025/06/10 22:06:41 Received task: 1 + 2
2025/06/10 22:06:41 Received task: 1 + 2
2025/06/10 22:06:41 Received task: 1 + 2
2025/06/10 22:06:41 Received task: 1 + 2
2025/06/10 22:06:41 Received task: 1 + 2
2025/06/10 22:06:41 Starting new worker with id 6
2025/06/10 22:06:41 Received task: 1 + 2
2025/06/10 22:06:41 Starting new worker with id 7
2025/06/10 22:06:41 Received task: 1 + 2
2025/06/10 22:06:41 Starting new worker with id 8
2025/06/10 22:06:41 Received task: 1 + 2
2025/06/10 22:06:41 Starting new worker with id 9
2025/06/10 22:06:41 Received task: 1 + 2
2025/06/10 22:06:41 Starting new worker with id 10
2025/06/10 22:06:41 Received task: 1 + 2
2025/06/10 22:06:41 Starting new worker with id 11
2025/06/10 22:06:41 Received task: 1 + 2
2025/06/10 22:06:41 Starting new worker with id 12
2025/06/10 22:06:41 Received task: 1 + 2
2025/06/10 22:06:41 Starting new worker with id 13
2025/06/10 22:06:41 Received task: 1 + 2
2025/06/10 22:06:41 Starting new worker with id 14
2025/06/10 22:06:41 Received task: 1 + 2
2025/06/10 22:06:41 Starting new worker with id 15
2025/06/10 22:06:41 Received task: 1 + 2
2025/06/10 22:06:41 Received task: 1 + 2
...
2025/06/10 22:07:05 Received task: 1 + 2
2025/06/10 22:07:38 Worker 2: idle timeout; exiting
2025/06/10 22:07:38 Worker 1: idle timeout; exiting
2025/06/10 22:07:38 Worker 1: stopped
2025/06/10 22:07:38 Worker 10: idle timeout; exiting
2025/06/10 22:07:38 Worker 10: stopped
2025/06/10 22:07:38 Worker 14: idle timeout; exiting
2025/06/10 22:07:38 Worker 14: stopped
2025/06/10 22:07:38 Worker 7: idle timeout; exiting
2025/06/10 22:07:38 Worker 7: stopped
2025/06/10 22:07:38 Worker 15: idle timeout; exiting
2025/06/10 22:07:38 Worker 15: stopped
2025/06/10 22:07:38 Worker 9: idle timeout; exiting
2025/06/10 22:07:38 Worker 9: stopped
2025/06/10 22:07:38 Worker 8: idle timeout; exiting
2025/06/10 22:07:38 Worker 6: idle timeout; exiting
2025/06/10 22:07:38 Worker 6: stopped
2025/06/10 22:07:38 Worker 8: stopped
2025/06/10 22:07:38 Worker 2: stopped
2025/06/10 22:07:38 Worker 13: idle timeout; exiting
2025/06/10 22:07:38 Worker 13: stopped


^C
2025/06/10 22:08:59 Shutting down server...
2025/06/10 22:08:59 Worker 11: stopped
2025/06/10 22:08:59 Worker 3: stopped
2025/06/10 22:08:59 Worker 4: stopped
2025/06/10 22:08:59 Worker 5: stopped
2025/06/10 22:08:59 Worker 12: stopped
```


unit-тесты:
```bash
$ go test -race ./internal/workerpool -v
```



