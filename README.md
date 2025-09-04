# Fast Track Data Ingestion Task


## Overview
This project implements a simple data ingestion pipeline using **Go**, **RabbitMQ**, and **MySQL**. The application publishes payment events into RabbitMQ, consumes them, and stores them in a MySQL database. Duplicate events are redirected into a separate table.


## Tech Stack
- **Go** (1.25)
- **RabbitMQ** (3-management)
- **MySQL** (8.0)
- **Docker & Docker Compose**


## Default values

- **MYSQL_HOST**=127.0.0.1
- **MYSQL_PORT**=3306
- **MYSQL_USER**=root
- **MYSQL_PASSWORD**=secret
- **MYSQL_DB**=fastdb
- **RABBIT_URL**=amqp://guest:guest@localhost:5672/
- **RABBIT_EXCHANGE**=payments
- **RABBIT_QUEUE**=payment_events_q
- **RABBIT_ROUTING_KEY**=payment.event


## How to Run

### 1. Clone the repository:
   ```bash
   git clone https://github.com/andrestrejos/fast-track-ingestion.git
   ```

### 2. Copy the environment file:
   ```bash
   cp configs/config.example.env .env
   ```

### 3. Start services (MySQL + RabbitMQ):
   ```bash
   docker compose up -d
   ```

### 4. Compile and Run the application:
   ```bash
   go mod tidy
   go run ./cmd/app
   ```


### The app will:

- Publish (1,1,10), (1,2,20), (2,3,20) in a separate goroutine.
- Consume and insert them into **payment_events**.
- Publish (1,1,10) again, causing a primary key violation.
- Redirect that duplicate to **skipped_messages**.

## Verification

### Connect to MySQL inside the container:
```bash
docker exec -it fast_mysql mysql -uroot -psecret fastdb
```
### Run:
```bash
SELECT * FROM payment_events ORDER BY payment_id;
SELECT * FROM skipped_messages ORDER BY payment_id;
```

- **RabbitMQ UI (optional)**: http://localhost:15672
 (guest/guest) → Queues → payment_events_q<br>
 → While running, Consumers = 1  → After processing, Ready = 0



## Notes:

- **No sleeps:** Publisher runs in its own goroutine; main waits on channels and a `WaitGroup`. This ensures deterministic sync without timing hacks.

- **Manual ACKs:** The consumer ACKs only after successful DB write; non-transient errors use Nack(false, false) to avoid requeue loops.

- **Duplicates:** Primary key on `payment_id` enforces uniqueness in **payment_events**. On violation, the event is inserted into **skipped_messages**.
We keep the same primary key in **skipped_messages** to record that an event with that `payment_id` was skipped, avoiding repetitive duplicates there as well.