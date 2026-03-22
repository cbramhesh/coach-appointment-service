# Coach Appointment Booking Service

A backend REST API built in Go that allows coaches to set their weekly availability and users to book 30-minute appointment slots.

## Tech Stack

- **Go** (1.25) — core language
- **chi** — lightweight HTTP router
- **MySQL 8** — relational database (via Docker)
- **go-playground/validator** — request validation
- **go-sql-driver/mysql** — MySQL driver

## How It Works

Coaches define recurring weekly availability (e.g., "Every Monday 9 AM–3 PM"). When a user requests available slots for a specific date, the system generates 30-minute windows on-the-fly, converts them to UTC, and filters out already-booked or past slots.

Double booking is prevented at the database level using a MySQL generated column with a unique index — even if two users try to book the same slot simultaneously, only one succeeds.

## Setup & Run

### Prerequisites
- Go 1.21+
- Docker

### 1. Clone the repo

```bash
git clone https://github.com/cbramhesh/coach-appointment-service.git
cd coach-appointment-service
```

### 2. Start MySQL via Docker

```bash
docker run --name coach-mysql \
  -e MYSQL_ROOT_PASSWORD=root123 \
  -e MYSQL_DATABASE=coach_booking \
  -p 3306:3306 \
  -d mysql:8
```

### 3. Create `.env` file

```bash
cp .env.example .env
```

Update `.env` with your values:
```
PORT=8080
DB_USER=root
DB_PASSWORD=root123
DB_HOST=localhost
DB_PORT=3306
DB_NAME=coach_booking
```

### 4. Run migrations

```bash
docker exec -i coach-mysql mysql -uroot -proot123 coach_booking < migrations/001_create_coaches.sql
docker exec -i coach-mysql mysql -uroot -proot123 coach_booking < migrations/002_create_availabilities.sql
docker exec -i coach-mysql mysql -uroot -proot123 coach_booking < migrations/003_create_users.sql
docker exec -i coach-mysql mysql -uroot -proot123 coach_booking < migrations/004_create_bookings.sql
docker exec -i coach-mysql mysql -uroot -proot123 coach_booking < migrations/005_seed_data.sql
```

### 5. Install dependencies & run

```bash
go mod tidy
go run cmd/server/main.go
```

You should see:
```
Connected to MySQL database
Server starting on :8080
```

## API Endpoints

| Method | Endpoint | Description |
|--------|----------|-------------|
| `GET` | `/health` | Health check |
| `POST` | `/coaches/availability` | Set weekly availability for a coach |
| `GET` | `/users/slots?coach_id=1&date=2026-03-23` | View available 30-min slots |
| `POST` | `/users/bookings` | Book an available slot |
| `GET` | `/users/bookings?user_id=1` | View upcoming bookings |
| `PATCH` | `/users/bookings/:id` | Cancel a booking |

## Example API Calls

### Set Coach Availability

```bash
curl -X POST http://localhost:8080/coaches/availability \
  -H "Content-Type: application/json" \
  -d '{"coach_id": 1, "day_of_week": 1, "start_time": "09:00", "end_time": "15:00"}'
```

### View Available Slots

```bash
curl "http://localhost:8080/users/slots?coach_id=1&date=2026-03-23"
```

Response:
```json
{
  "success": true,
  "data": {
    "coach_id": 1,
    "date": "2026-03-23",
    "time_zone": "America/New_York",
    "slots": [
      { "start_time": "2026-03-23T13:00:00Z", "end_time": "2026-03-23T13:30:00Z" },
      { "start_time": "2026-03-23T13:30:00Z", "end_time": "2026-03-23T14:00:00Z" }
    ]
  }
}
```

### Book a Slot

```bash
curl -X POST http://localhost:8080/users/bookings \
  -H "Content-Type: application/json" \
  -d '{"user_id": 1, "coach_id": 1, "start_time": "2026-03-23T14:00:00Z"}'
```

### View Upcoming Bookings

```bash
curl "http://localhost:8080/users/bookings?user_id=1"
```

### Cancel a Booking

```bash
curl -X PATCH http://localhost:8080/users/bookings/1 \
  -H "Content-Type: application/json" \
  -d '{"user_id": 1}'
```

## Project Structure

```
coach-appointment-service/
├── cmd/server/main.go          # Entry point
├── internal/
│   ├── config/                 # Environment config
│   ├── apperror/               # Custom error types
│   ├── model/                  # Database models
│   ├── dto/                    # Request/response DTOs
│   ├── repository/             # SQL queries (data layer)
│   ├── service/                # Business logic
│   ├── handler/                # HTTP handlers
│   └── router/                 # Route definitions
└── migrations/                 # SQL migration files
```

## Design Decisions

**Slot Generation**: Slots are computed on-the-fly from availability windows rather than being pre-stored. This keeps the data model simple and avoids stale slot data when availability changes.

**Timezone Handling**: Coach availability is stored as local `TIME` values. Bookings are stored in UTC. Timezone conversion happens during slot generation using Go's `time.LoadLocation`.

**Double Booking Prevention**: A MySQL generated column (`active_slot`) creates a composite key of `coach_id + start_time` for confirmed bookings. A unique index on this column prevents two confirmed bookings for the same coach at the same time. Cancelled bookings have a NULL `active_slot`, so they don't block rebooking.

**Error Handling**: Custom `AppError` type maps application errors to HTTP status codes (`400`, `404`, `409`, `500`). All responses follow a consistent JSON envelope with `success` boolean and either `data` or `error` fields.

## Seed Data

The migrations include seed data for testing:

| Type | Name | Timezone |
|------|------|----------|
| Coach | Coach Smith | America/New_York |
| Coach | Coach Patel | Asia/Kolkata |
| User | Alice | America/New_York |
| User | Bob | Asia/Kolkata |

## AI Assistance

I used AI (Claude) as a design advisor and learning companion while building this project. The AI helped me with:
- Breaking down the problem into components
- Designing the database schema and understanding trade-offs
- Understanding Go patterns (layered architecture, error handling)
- Debugging issues (e.g., `time.Time` vs `string` for MySQL TIME columns)

The actual code was written by me based on the AI's guidance and suggestions.
