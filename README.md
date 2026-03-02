# TripShare — Shared Trip Expense Manager

A production-ready web application for managing shared trip expenses among a group of friends. Built with Go, SQLite, and a clean modern UI.

## Features

- **Group Management** — Add, edit, and remove group members dynamically
- **Trip Management** — Create trips with custom participants and costs
- **Payment Tracking** — Record who paid what on each trip
- **Automatic Balancing** — Calculates equal shares, overpayments, and underpayments
- **Persistent Balances** — Balances carry over across trips
- **Settlement Suggestions** — Shows who owes whom and how much
- **Clean UI** — Modern, responsive web interface

## Project Structure

```
.
├── cmd/server/main.go              # Application entry point
├── internal/
│   ├── models/models.go            # Data models and request/response types
│   ├── database/database.go        # SQLite + GORM initialization
│   ├── repository/repository.go    # Data access layer
│   ├── service/
│   │   ├── member.go               # Member business logic
│   │   ├── trip.go                 # Trip business logic
│   │   ├── payment.go              # Payment business logic
│   │   ├── balance.go              # Balance calculation engine
│   │   └── balance_test.go         # Unit tests
│   ├── handler/
│   │   ├── util.go                 # HTTP utilities
│   │   ├── member.go               # Member endpoints
│   │   ├── trip.go                 # Trip endpoints
│   │   ├── payment.go              # Payment endpoints
│   │   └── balance.go              # Balance endpoint
│   └── router/router.go            # Route registration
├── web/
│   ├── templates/index.html        # Single-page UI
│   └── static/
│       ├── style.css               # Styles
│       └── app.js                  # Client-side logic
├── Dockerfile
├── docker-compose.yml
├── go.mod
└── README.md
```

## Quick Start

### Prerequisites

- Go 1.23+ (with CGO enabled for SQLite)
- GCC (for SQLite compilation)

### Run Locally

```bash
# Install dependencies
go mod tidy

# Run the server
go run ./cmd/server

# Open in browser
# http://localhost:8080
```

### Run with Docker

```bash
docker compose up --build
# Open http://localhost:8080
```

### Run Tests

```bash
go test ./internal/service/ -v
```

## API Endpoints

### Members

```bash
# Create a member
curl -X POST http://localhost:8080/members \
  -H "Content-Type: application/json" \
  -d '{"name": "Alice"}'

# List all members
curl http://localhost:8080/members

# Update a member
curl -X PUT http://localhost:8080/members/1 \
  -H "Content-Type: application/json" \
  -d '{"name": "Alice Smith"}'

# Delete a member
curl -X DELETE http://localhost:8080/members/1
```

### Trips

```bash
# Create a trip
curl -X POST http://localhost:8080/trips \
  -H "Content-Type: application/json" \
  -d '{"name": "Beach Vacation", "total_cost": 1000, "date": "2025-07-15", "member_ids": [1,2,3,4]}'

# List all trips
curl http://localhost:8080/trips

# Get trip details with breakdown
curl http://localhost:8080/trips/1

# Update a trip
curl -X PUT http://localhost:8080/trips/1 \
  -H "Content-Type: application/json" \
  -d '{"name": "Beach Vacation 2025", "total_cost": 1200, "member_ids": [1,2,3,4]}'

# Delete a trip
curl -X DELETE http://localhost:8080/trips/1
```

### Payments

```bash
# Add a payment
curl -X POST http://localhost:8080/trips/1/payments \
  -H "Content-Type: application/json" \
  -d '{"member_id": 1, "amount": 1000}'

# Update a payment
curl -X PUT http://localhost:8080/trips/1/payments/1 \
  -H "Content-Type: application/json" \
  -d '{"amount": 800}'

# Delete a payment
curl -X DELETE http://localhost:8080/trips/1/payments/1
```

### Balances

```bash
# Get all balances and settlement suggestions
curl http://localhost:8080/balances
```

## Example Scenario

```bash
# 1. Add 4 members
curl -s -X POST http://localhost:8080/members -H "Content-Type: application/json" -d '{"name":"Alice"}'
curl -s -X POST http://localhost:8080/members -H "Content-Type: application/json" -d '{"name":"Bob"}'
curl -s -X POST http://localhost:8080/members -H "Content-Type: application/json" -d '{"name":"Carol"}'
curl -s -X POST http://localhost:8080/members -H "Content-Type: application/json" -d '{"name":"Dave"}'

# 2. Create a trip for 1000 EUR with all 4 members
curl -s -X POST http://localhost:8080/trips \
  -H "Content-Type: application/json" \
  -d '{"name":"Summer Holiday","total_cost":1000,"date":"2025-08-01","member_ids":[1,2,3,4]}'

# 3. Alice pays the entire 1000 EUR
curl -s -X POST http://localhost:8080/trips/1/payments \
  -H "Content-Type: application/json" \
  -d '{"member_id":1,"amount":1000}'

# 4. Check balances — Alice: +750, Bob/Carol/Dave: -250 each
curl -s http://localhost:8080/balances | python3 -m json.tool
```

## Database

Uses SQLite with GORM for ORM. The database file is created automatically at startup. Tables are auto-migrated. WAL mode and foreign keys are enabled for reliability.

## Architecture

The application follows a layered architecture:

1. **Handler** — HTTP request parsing, validation, response writing
2. **Service** — Business logic, balance calculations, cross-cutting concerns
3. **Repository** — Database access via GORM
4. **Models** — Shared data structures

Balances are recomputed from all trip/payment data after every mutation, ensuring consistency.
