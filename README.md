# Fast Track Data Ingestion Task


## Overview
This project implements a simple data ingestion pipeline using **Go**, **RabbitMQ**, and **MySQL**. The application publishes payment events into RabbitMQ, consumes them, and stores them in a MySQL database. Duplicate events are redirected into a separate table.


## Tech Stack
- **Go** (1.25)
- **RabbitMQ** (3-management)
- **MySQL** (8.0)
- **Docker & Docker Compose**


## Setup
```bash
git clone <repo>
cd fast-track-ingestion
cp configs/config.example.env .env
docker compose up -d
go run ./cmd/app