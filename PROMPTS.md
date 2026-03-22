# AI Prompts & Responses Used During Development

---

## Prompt 1: Breaking Down the Problem

**My Prompt:**
> Hey, I've been given a backend assignment to build an appointment booking system where users can book 30-minute slots with coaches. Before jumping into coding, I want to first think through the design properly. Can you help me break down the problem into smaller components, suggest a clean backend architecture, recommend what entities/tables I should have, and highlight tricky parts like slot generation and double booking?

**AI Response:**

The system connects **coaches** with **users** for 30-minute appointment slots. At its core, there are four major capabilities:

| # | Capability | Who | Complexity |
|---|-----------|-----|------------|
| 1 | Set weekly availability | Coach | Low |
| 2 | View available 30-min slots for a coach on a date | User | **Medium** — requires slot generation logic |
| 3 | Book a slot | User | **High** — must prevent double booking & race conditions |
| 4 | View upcoming bookings | User | Low |

### Sub-problems worth isolating

1. **Availability modeling** — How do we store "I'm free Mon 10 AM–3 PM"?
2. **Slot generation** — How do we turn a time range into discrete 30-min windows?
3. **Booking & conflict detection** — How do we guarantee no two users book the same slot?
4. **Timezone handling** — Coach and user may be in different timezones.

### Proposed Architecture

A clean **layered architecture** works well for this scope:

```
┌───────────────────────────────────────────┐
│              API / Routes Layer           │  ← Routes
│  (input validation, HTTP concerns)        │
├───────────────────────────────────────────┤
│            Service / Business Layer       │  ← Core logic lives here
│  (slot generation, booking rules)         │
├───────────────────────────────────────────┤
│           Repository / Data Layer         │  ← DB queries, transactions
│  (SQL queries, row-level locking)         │
├───────────────────────────────────────────┤
│              Database (MySQL)             │
└───────────────────────────────────────────┘
```

**Why this works:**
- **Routes** only handle request parsing, validation, and response formatting.
- **Services** contain all business logic (slot generation, conflict checks) and are independently testable.
- **Repositories** encapsulate raw SQL calls, making it easy to swap databases or add caching later.

### Entity/Table Design

You need 4 tables:
- **coaches** — coach info + timezone
- **users** — user info + email
- **availabilities** — coach's recurring weekly availability windows (day + start/end TIME)
- **bookings** — actual confirmed/cancelled appointments (DATETIME in UTC)

### Tricky Parts

The core complexity lives in two areas:
1. **Slot generation** — turning weekly availability rules + a specific date into concrete UTC time windows, accounting for timezones and existing bookings.
2. **Concurrency-safe booking** — using database constraints to prevent double bookings, even under concurrent requests.

---

## Prompt 2: Database Schema Design

**My Prompt:**
> I'm planning to use a relational database (MySQL). Can you help me design the schema for coaches, users, availability, and bookings? I want proper relationships, constraints, and fields that will make slot booking efficient. Also, how should I structure availability so that generating 30-minute slots becomes easy?

**AI Response:**

### Why TIME-Based Windows (Not Pre-Generated Slots)

There are two ways to model availability:

| Approach | How it works | Verdict |
|----------|-------------|---------|
| **A) Store time windows** ✅ | Coach says "Monday 09:00–15:00". We store that range. At query time, we split it into 30-min slots and subtract bookings. | Simple, flexible, no stale data |
| B) Pre-generate every slot | When coach sets availability, immediately INSERT rows for every 30-min slot | Storage grows fast, must regenerate on every change, stale data risk |

**We go with Approach A.** The `availabilities` table stores windows; slots are computed at query time.

### Full Schema

**coaches:**
```sql
CREATE TABLE coaches (
    id INT AUTO_INCREMENT PRIMARY KEY,
    name VARCHAR(100) NOT NULL,
    timezone VARCHAR(50) NOT NULL DEFAULT 'UTC',
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP
);
```

**Why `timezone` on the coach?** Availability times like "09:00–15:00" are meaningless without knowing *whose* 09:00. This column anchors those times.

**availabilities:**
```sql
CREATE TABLE availabilities (
    id INT AUTO_INCREMENT PRIMARY KEY,
    coach_id INT NOT NULL,
    day_of_week TINYINT NOT NULL CHECK (day_of_week BETWEEN 0 AND 6),
    start_time TIME NOT NULL,
    end_time TIME NOT NULL,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    CONSTRAINT fk_avail_coach FOREIGN KEY (coach_id) REFERENCES coaches(id) ON DELETE CASCADE,
    CONSTRAINT chk_valid_window CHECK (end_time > start_time),
    CONSTRAINT uq_coach_availability UNIQUE (coach_id, day_of_week, start_time, end_time)
);
```

> **`TIME` (without timezone)** is intentional. These are coach-local times. The coach's `timezone` column provides the context. `start_time = '09:00'` always means 9 AM in the coach's timezone.

**bookings:**
```sql
CREATE TABLE bookings (
    id INT AUTO_INCREMENT PRIMARY KEY,
    user_id INT NOT NULL,
    coach_id INT NOT NULL,
    start_time DATETIME NOT NULL,
    end_time DATETIME NOT NULL,
    status VARCHAR(20) NOT NULL DEFAULT 'confirmed',
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    active_slot VARCHAR(100) GENERATED ALWAYS AS (
        CASE WHEN status = 'confirmed'
             THEN CONCAT(coach_id, '_', start_time)
             ELSE NULL
        END
    ) STORED,
    UNIQUE INDEX idx_no_double_booking (active_slot)
);
```

The `active_slot` generated column is a MySQL workaround for preventing double booking. When status is 'confirmed', it generates a unique string like `"1_2026-03-23 14:00:00"`. The unique index rejects duplicates. When status is 'cancelled', it's NULL — MySQL unique indexes ignore NULLs, so cancelled bookings don't block rebooking.

---

## Prompt 3: Slot Generation Logic

**My Prompt:**
> I'm trying to figure out how to generate 30-minute slots from a coach's availability. For example, if a coach is available from 9:00 to 12:00, how should I break this into 30-minute slots, exclude already booked slots, and make this efficient?

**AI Response:**

Here's a concrete worked example:

**Setup:**
- Coach 1 is in `America/New_York`
- Available Monday 09:00–12:00 (local time)
- One existing booking: `2025-10-27T14:00:00Z` (which is Mon 10:00 AM ET)

**Query: "Show Coach 1's available slots for Monday Oct 27"**

```
Step 1 — Get coach timezone
  → 'America/New_York'

Step 2 — Get availability rules for Monday (day_of_week = 1)
  → [{ start_time: '09:00', end_time: '12:00' }]

Step 3 — Generate all possible 30-min slots
  09:00→09:30, 09:30→10:00, 10:00→10:30, 10:30→11:00, 11:00→11:30, 11:30→12:00

Step 4 — Convert to UTC using coach's timezone + date
  09:00 ET = 13:00 UTC  (EDT, UTC-4)
  09:30 ET = 13:30 UTC
  10:00 ET = 14:00 UTC
  ...

Step 5 — Fetch existing confirmed bookings for that date range
  → ['2025-10-27T14:00Z']  (the 10:00 AM ET slot)

Step 6 — Subtract booked from generated
  Available = [13:00, 13:30, 14:30, 15:00, 15:30] UTC
  (14:00 UTC removed — it's booked)

Step 7 — Return to user
```

**Why compute on-the-fly instead of pre-generating?**
- Storage efficiency: 1 availability row vs potentially 12+ slot rows per day
- Flexibility: if a coach changes their availability, no orphaned slot records
- Simplicity: no background job needed

**Edge cases to handle:**
- Past slots: if the date is today, filter out slots already in the past
- Timezone DST transitions: use proper timezone libraries
- Slot must start on :00 or :30 boundary

---

## Prompt 4: API Design

**My Prompt:**
> Now I want to design the APIs properly. Can you help me define endpoints, request/response format, validation rules, and error handling? I want them to be RESTful and clean. Also mention what status codes I should return.

**AI Response:**

### Design Principles
- **Consistent envelope** — every response follows the same shape
- **Resource-oriented URLs** — nouns, not verbs (`/bookings` not `/createBooking`)
- **Proper HTTP verbs** — POST to create, GET to read, PATCH to partial-update
- **Meaningful status codes** — not just 200 and 500

### Response Envelope

```json
// Success
{ "success": true, "data": { ... } }

// Error
{ "success": false, "error": { "code": "SLOT_ALREADY_BOOKED", "message": "This slot is already booked" } }
```

### Endpoints

**POST /coaches/availability** — `201 Created`
```json
{ "coach_id": 1, "day_of_week": 1, "start_time": "09:00", "end_time": "15:00" }
```

**GET /users/slots?coach_id=1&date=2025-10-28** — `200 OK`
Returns computed available 30-min slots in UTC.

**POST /users/bookings** — `201 Created`
```json
{ "user_id": 101, "coach_id": 1, "start_time": "2025-10-28T13:00:00Z" }
```
Client sends only `start_time`. Server computes `end_time = start_time + 30 min`.

**GET /users/bookings?user_id=101** — `200 OK`
Returns upcoming confirmed bookings.

**PATCH /users/bookings/:id** — `200 OK`
Soft-cancel a booking (sets status to 'cancelled').

### Status Codes

| Code | When |
|------|------|
| 200 | Successful GET, successful cancel |
| 201 | Resource created (POST) |
| 400 | Validation errors, bad format, slot not on boundary |
| 403 | User trying to cancel someone else's booking |
| 404 | Coach/user/booking not found |
| 409 | Slot already booked, duplicate availability |
| 500 | Unexpected server error |

### Error Codes
`VALIDATION_ERROR`, `INVALID_SLOT`, `SLOT_IN_PAST`, `COACH_NOT_FOUND`, `BOOKING_NOT_FOUND`, `FORBIDDEN`, `SLOT_ALREADY_BOOKED`, `DUPLICATE_AVAILABILITY`, `ALREADY_CANCELLED`

---

## Prompt 5: Go Project Structure

**My Prompt:**
> I'll be implementing this using Golang with chi router. Can you help me structure the project properly with controllers, services, repositories, and DTOs? Also suggest package structure and best practices.

**AI Response:**

### Directory Layout

```
coach-appointment-service/
├── cmd/server/
│   └── main.go              # Entry point — wires everything, starts server
├── internal/
│   ├── config/config.go     # Env-based configuration
│   ├── handler/             # HTTP handlers ("controllers" in Go)
│   │   ├── coach_handler.go
│   │   ├── user_handler.go
│   │   └── response.go      # Shared response helpers
│   ├── service/             # Business logic layer
│   │   ├── availability_service.go
│   │   ├── slot_service.go   # Core slot generation algorithm
│   │   └── booking_service.go
│   ├── repository/          # Data access layer
│   │   ├── coach_repo.go
│   │   ├── availability_repo.go
│   │   └── booking_repo.go
│   ├── model/               # DB row structs
│   ├── dto/                 # Request/Response DTOs
│   ├── router/router.go     # Chi route definitions
│   └── apperror/errors.go   # Custom error types
├── migrations/              # SQL migration files
├── go.mod
├── .env.example
└── README.md
```

**Why `internal/`?** Go's `internal` directory enforces compile-time encapsulation — packages outside can't import from it.

**Why separate `model/` and `dto/`?**
- `model` maps 1:1 to database rows
- `dto` shapes what the API receives/returns — prevents leaking DB internals

### Data Flow

```
Client → Handler → Service → Repository → MySQL
                                          ↓
Client ← Handler ← Service ← Repository ←
```

**Key rule**: Each layer only talks to the one directly below it. Handlers never touch the DB. Repositories never know about HTTP.

### Wiring in main.go

```go
func main() {
    cfg := config.LoadConfig()
    db, _ := sql.Open("mysql", cfg.DSN)

    // Repos → Services → Handlers → Router
    coachRepo := repository.NewCoachRepo(db)
    availRepo := repository.NewAvailabilityRepo(db)
    bookingRepo := repository.NewBookingRepo(db)

    availSvc := service.NewAvailabilityService(availRepo, coachRepo)
    slotSvc := service.NewSlotService(coachRepo, availRepo, bookingRepo)
    bookingSvc := service.NewBookingService(bookingRepo, coachRepo, slotSvc)

    coachHandler := handler.NewCoachHandler(availSvc)
    userHandler := handler.NewUserHandler(slotSvc, bookingSvc)

    r := router.NewRouter(coachHandler, userHandler)
    http.ListenAndServe(":"+cfg.Port, r)
}
```

---

## Prompt 6: Double Booking Prevention

**My Prompt:**
> I'm concerned about double booking when two users try to book the same slot at the same time. Can you explain how this problem happens and how to solve it using database-level solutions?

**AI Response:**

### The Race Condition

Two users (Alice & Bob) try to book the **same slot** at almost the same instant:

```
Time    Alice's Request                    Bob's Request
─────   ─────────────────────              ─────────────────────
t1      SELECT: "Is 10:00 free?"
t2      Result: ✅ Yes, it's free
t3                                         SELECT: "Is 10:00 free?"
t4                                         Result: ✅ Yes, it's free
t5      INSERT booking (10:00)
t6      ✅ Alice booked!
t7                                         INSERT booking (10:00)
t8                                         ✅ Bob also booked!  ← 💥 DOUBLE BOOKING
```

**Why it happens**: Between the SELECT check and the INSERT, there's a gap where both requests see the slot as free. This is a **TOCTOU** (Time Of Check, Time Of Use) bug.

> Application-level checks alone will **always** have this race window. You **must** use database-level mechanisms.

### Solution: Generated Column + Unique Index (MySQL)

Since MySQL doesn't support partial unique indexes, we use a generated column:

```sql
active_slot VARCHAR(100) GENERATED ALWAYS AS (
    CASE WHEN status = 'confirmed'
         THEN CONCAT(coach_id, '_', start_time)
         ELSE NULL
    END
) STORED,
UNIQUE INDEX idx_no_double_booking (active_slot)
```

**How it works:**
- For confirmed bookings: `active_slot` = `"1_2026-03-23 14:00:00"` → unique
- For cancelled bookings: `active_slot` = NULL → ignored by unique index
- If two concurrent inserts hit the same slot, second one fails with `Duplicate entry`

**In Go, catch the error:**
```go
result, err := s.bookingRepo.Create(ctx, booking)
if err != nil {
    if strings.Contains(err.Error(), "Duplicate entry") {
        return nil, apperror.Conflict("SLOT_ALREADY_BOOKED",
            "This time slot is no longer available")
    }
    return nil, err
}
```

### Two Layers of Defense

1. **Application-level check** (optimistic): Before inserting, call `SlotService.GetAvailableSlots` to verify the slot is actually available. Catches 99% of conflicts.

2. **Database unique constraint** (pessimistic): The generated column + unique index catches the remaining 1% — concurrent requests that both pass the app check.

---

## How I Used AI

I used AI (Claude) as a **learning companion and design advisor** throughout this project. Here's my approach:

1. **Design first**: I asked for help understanding the problem, entity relationships, and edge cases before writing any code
2. **Architecture guidance**: I got suggestions on project structure and Go idioms but wrote the actual code myself
3. **Incremental building**: I built the project feature-by-feature, committing at each step
4. **Debugging help**: When I hit issues (like the time parsing bug), I shared errors and worked through fixes
5. **Understanding patterns**: I asked for explanations of WHY certain patterns work (like the generated column trick) so I can explain them in interviews

The AI helped me with design decisions and pointed me toward Go patterns I wasn't familiar with, but the actual code was written and structured by me based on that guidance.
