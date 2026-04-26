# Kitty Party App — Technical Documentation

A Go REST API that models the lifecycle of Indian-style "kitty party" savings pools: groups of members contribute a fixed amount each month, and one member hosts and receives the pooled amount each cycle. This document describes the architecture, domains, data model, and HTTP surface — not how to run the app (see [README.md](README.md#L1) for that).

---

## 1. Runtime Stack

| Concern           | Choice                                    |
| ----------------- | ----------------------------------------- |
| Language          | Go 1.25.6 ([go.mod](go.mod#L3))           |
| HTTP framework    | Gin (`github.com/gin-gonic/gin`)          |
| Database driver   | pgx v5 (`github.com/jackc/pgx/v5`)        |
| Database          | PostgreSQL (UUID PKs, `gen_random_uuid()`)|
| Auth              | JWT HS256 (`golang-jwt/jwt/v5`)           |
| Structured logs   | Uber Zap                                  |
| Config / env      | `joho/godotenv`                           |
| Validation        | `go-playground/validator` (Gin `binding`) |
| Container         | Dockerfile + `docker-compose.yml`         |

Entry point: [cmd/api/main.go](cmd/api/main.go#L1) wires all domains (Repository → Service → Handler) and hands the composed router to a graceful HTTP server.

---

## 2. Architectural Layering

The codebase follows a strict **clean-architecture / hexagonal** split. Every bounded context under [internal/](internal/) has the same five files:

```
internal/<domain>/
├── model.go                # Entities + request/response DTOs with binding tags
├── repository.go           # Repository interface (persistence port)
├── postgres_repository.go  # pgx-backed adapter
├── service.go              # Service interface + business rules
└── handler.go              # Gin handler + RegisterRoutes
```

Dependency direction: `handler → service → repository`. Interfaces live in the consumer package; concrete DB adapters depend on the interface. Cross-domain dependencies are expressed as **provider adapters** (see §4).

### Request lifecycle

1. Gin engine ([internal/router/router.go](internal/router/router.go#L30)) applies global middleware — Recovery, structured Logger, CORS.
2. Handler binds & validates JSON (`c.ShouldBindJSON`) with struct tags like `binding:"required,gt=0"`.
3. Service enforces business invariants and calls the repository.
4. Repository runs parameterized SQL inside `context.WithTimeout`; multi-row writes use `pgxpool` transactions with deferred rollback.
5. Errors are wrapped via [apperrors](internal/apperrors/errors.go#L1) (HTTP status + message) and rendered by [response.Error](internal/response/response.go).

---

## 3. Bounded Contexts (Domains)

### 3.1 `auth` — JWT issuance
- [internal/auth/handler.go](internal/auth/handler.go#L1): `POST /api/v1/auth/token` issues an HS256 JWT used by protected routes.
- [internal/middleware/auth.go](internal/middleware/auth.go#L1): `AuthMiddleware()` validates `Authorization: Bearer <jwt>` and writes `user_id` into the Gin context.
- Also shipped: [cmd/tokengen/main.go](cmd/tokengen/main.go#L1) — standalone CLI for generating a test JWT.

### 3.2 `member` — user identity
- Entity: [Member](internal/member/model.go#L7) (`id`, `name`, `email`, `phone`).
- Public CRUD under `/api/v1/members`.
- Stored in `users` table (see [V1__create_users_table.sql](db/scripts/V1__create_users_table.sql)).

### 3.3 `profile` — extended member data
- Entity: [MemberProfile](internal/profile/model.go#L7) — DOB, picture, address, occupation, about.
- Rich read model: [ProfileSummary](internal/profile/model.go#L58) aggregates groups + `total_paid` + `pending_dues` across domains.
- Uses [MemberChecker](internal/profile/member_checker.go) port to validate member existence without importing the `member` package directly — clean dependency boundary.

### 3.4 `group` — kitty party pool + memberships
Two closely related sub-domains colocated in the same package:
- [Group](internal/group/model.go#L10): pool config (`monthly_amount`, `duration`, `start_date`, `created_by`). Hard cap `MaxGroupMembers = 10`.
- [GroupMembership](internal/group/model.go#L39): a member's participation with `status` (ACTIVE/PENDING/REJECTED) and `role` (MEMBER/ADMIN).
- Separate [membership_repository.go](internal/group/membership_repository.go) / [membership_service.go](internal/group/membership_service.go) / [membership_handler.go](internal/group/membership_handler.go) keep the two concerns split even while sharing a package.
- Routes: `/groups`, `/groups/:id/members`, `/groups-member-counts`.

### 3.5 `kittycycle` — scheduled rotation
This is the heart of the rotation logic. See [internal/kittycycle/model.go](internal/kittycycle/model.go#L1).

- [KittyCycle](internal/kittycycle/model.go#L8): the parent rotation record — `total_cycles`, `monthly_amount`, `pool_amount`, `start_date`, `end_date`, status (PENDING/ACTIVE/COMPLETED/CANCELLED).
- [KittySchedule](internal/kittycycle/model.go#L27): one row per month = one host + one scheduled party, with `cycle_number`, `host_member_id`, `scheduled_date`, `due_date`, `pool_amount`, and **`venue`** (optional — where the kitty was held).
- Atomic creation: [CreateCycleWithSchedule](internal/kittycycle/postgres_repository.go#L23) inserts the cycle and all month rows inside a single pgx transaction; duplicate `cycle_number` or `host_member_id` inside the same cycle is caught via `uq_` constraints and mapped to HTTP 409.
- Business invariants enforced in [service.go](internal/kittycycle/service.go#L37):
  - Only group admin/creator can schedule or cancel.
  - `end_date > start_date`.
  - `len(schedule) ≤ group.Duration`.
  - No duplicate `cycle_number` or host within a cycle.
  - `due_date ≤ scheduled_date`.
  - Each host must actually be a group member.
- Cross-domain dependency is inverted via the [GroupInfoProvider](internal/kittycycle/group_provider_adapter.go) port — the cycle package never imports `group` directly; it receives an adapter at wire time.

### 3.6 `contribution` — monthly dues + payments
- [ContributionDue](internal/contribution/model.go#L6): per-member, per-cycle obligation (`PENDING`/`PAID`/`OVERDUE`).
- [ContributionPayment](internal/contribution/model.go#L23): a real transaction against a due (`UPI`/`CASH`/`BANK`, `SUCCESS`/`FAILED`).
- `POST /contributions/dues/generate` fans out dues for every ACTIVE group member for a given cycle.

### 3.7 `payout` — pool disbursement
- [PayoutSchedule](internal/payout/model.go#L8): who receives the pooled amount for a given cycle.
- [Disbursement](internal/payout/model.go#L25): the actual money transfer settling a scheduled payout.
- [PayoutSummary](internal/payout/model.go#L38) joins schedule + disbursement for convenient reads.

---

## 4. Cross-Domain Dependencies (Provider Adapters)

To keep packages decoupled, dependent domains define their own small interface and accept an adapter that wraps another domain's repository:

| Consumer      | Port                                                                  | Adapter file                                                     |
| ------------- | --------------------------------------------------------------------- | ---------------------------------------------------------------- |
| `contribution`| `GroupMemberProvider`                                                 | [provider_adapter.go](internal/contribution/provider_adapter.go) |
| `payout`      | `GroupInfoProvider`                                                   | [group_provider_adapter.go](internal/payout/group_provider_adapter.go) |
| `kittycycle`  | `GroupInfoProvider` (needs `GetGroup`, `IsAdminOrCreator`, `IsMember`)| [group_provider_adapter.go](internal/kittycycle/group_provider_adapter.go) |
| `profile`     | `MemberChecker`                                                       | [member_checker.go](internal/profile/member_checker.go)          |

This means every domain can be unit-tested with fake adapters, and swapping the backing store (e.g., adding Mongo) touches only adapters + `postgres_repository.go`.

---

## 5. Data Model & Migrations

Flyway-style versioned SQL under [db/scripts/](db/scripts/). All DDL runs inside `BEGIN; ... COMMIT;` blocks and uses `IF NOT EXISTS` guards for idempotency.

| Version | Purpose                                                                 |
| ------- | ----------------------------------------------------------------------- |
| V0      | Database bootstrap                                                      |
| V1      | `users`                                                                 |
| V2      | `groups`                                                                |
| V3      | `group_memberships` (join table, unique (group_id, member_id))          |
| V4      | Add `status` to `group_memberships`                                     |
| V5      | `contribution_due`                                                      |
| V6      | `contribution_payment`                                                  |
| V7      | Add `role` to `group_memberships`                                       |
| V8      | `member_profiles` (1:1 with users)                                      |
| V9      | `payout_schedule`                                                       |
| V10     | `disbursement`                                                          |
| V11     | `kitty_cycle`                                                           |
| V12     | `kitty_schedule` (uq on (cycle_id, cycle_number) & (cycle_id, host))    |
| V13     | Add `venue` to `kitty_schedule`                                         |

Seed/dummy data lives under [db/data/](db/data/) and [db/seeds/](db/seeds/) — numbered separately from schema migrations.

### Conventions
- Every PK: `UUID DEFAULT gen_random_uuid()`.
- Monetary columns: `NUMERIC(12,2) CHECK (> 0)`.
- Timestamps: `TIMESTAMP WITH TIME ZONE DEFAULT NOW()`; `updated_at` is nullable and set by `UPDATE` statements.
- Referential integrity: `ON DELETE CASCADE` for child tables owned by a parent (e.g., `kitty_schedule → kitty_cycle`), `ON DELETE RESTRICT` for identity references (e.g., `host_member_id → users`).
- Enum-like fields are modeled as `VARCHAR(20)` with `CHECK (... IN (...))`, not as PostgreSQL enums (easier to evolve with new migrations).

---

## 6. HTTP Surface

All routes are prefixed `/api/v1`. Auth column marks routes wrapped by `middleware.AuthMiddleware()`.

| Method | Path                                         | Auth | Purpose                                  |
| ------ | -------------------------------------------- | ---- | ---------------------------------------- |
| GET    | `/health`                                    | –    | Liveness check                           |
| POST   | `/api/v1/auth/token`                         | –    | Issue JWT                                |
| GET    | `/api/v1/members`                            | –    | List members                             |
| POST   | `/api/v1/members`                            | –    | Create member                            |
| GET    | `/api/v1/members/:id`                        | –    | Get member                               |
| PUT    | `/api/v1/members/:id`                        | –    | Update member                            |
| DELETE | `/api/v1/members/:id`                        | –    | Delete member                            |
| GET    | `/api/v1/members/:id/profile`                | –    | Get extended profile                     |
| PUT    | `/api/v1/members/:id/profile`                | ✔    | Upsert extended profile                  |
| GET    | `/api/v1/members/:id/profile/summary`        | –    | Aggregated profile (groups + financials) |
| GET    | `/api/v1/groups`                             | –    | List groups                              |
| POST   | `/api/v1/groups`                             | –    | Create group                             |
| GET    | `/api/v1/groups/:id`                         | –    | Get group                                |
| PUT    | `/api/v1/groups/:id`                         | –    | Update group                             |
| DELETE | `/api/v1/groups/:id`                         | –    | Delete group                             |
| GET    | `/api/v1/groups-member-counts`               | –    | Count members across all groups          |
| GET    | `/api/v1/groups/:id/members`                 | –    | List members in a group                  |
| POST   | `/api/v1/groups/:id/members`                 | ✔    | Add member to group                      |
| GET    | `/api/v1/groups/:id/members/count`           | –    | Count members in one group               |
| POST   | `/api/v1/contributions/dues/generate`        | –    | Fan out monthly dues                     |
| GET    | `/api/v1/contributions/groups/:groupID/dues` | –    | List dues for a group                    |
| POST   | `/api/v1/contributions/payments`             | –    | Record a payment                         |
| POST   | `/api/v1/payouts/schedule`                   | ✔    | Schedule a payout                        |
| POST   | `/api/v1/payouts/disburse`                   | ✔    | Record a disbursement                    |
| GET    | `/api/v1/payouts/groups/:groupID`            | ✔    | List payouts for a group                 |
| GET    | `/api/v1/payouts/recipients/:memberID`       | ✔    | List payouts for a member                |
| GET    | `/api/v1/payouts/:id`                        | ✔    | Get payout                               |
| PATCH  | `/api/v1/payouts/:id/cancel`                 | ✔    | Cancel a payout                          |
| POST   | `/api/v1/kitty-cycles/schedule`              | ✔    | Create cycle + all monthly rows atomically |
| GET    | `/api/v1/kitty-cycles/groups/:groupID`       | ✔    | List cycles for a group                  |
| GET    | `/api/v1/kitty-cycles/:id`                   | ✔    | Get cycle with full schedule             |
| PATCH  | `/api/v1/kitty-cycles/:id/cancel`            | ✔    | Cancel a PENDING/ACTIVE cycle            |

### Response envelope
All successful responses are wrapped by [response.OK / Created / NoContent](internal/response/response.go); all errors are rendered as `{ "success": false, "error": "…" }` with the HTTP status carried on the `apperrors.AppError`.

---

## 7. Configuration

Loaded in [internal/config/config.go](internal/config/config.go#L21) from env (with `.env` fallback via `godotenv`):

| Variable      | Default        | Purpose                          |
| ------------- | -------------- | -------------------------------- |
| `APP_NAME`    | kitty-party-app| Shown in logs                    |
| `APP_ENV`     | development    | Switches Zap logger profile      |
| `APP_PORT`    | 8080           | HTTP listen port                 |
| `DB_USER`     | postgres       | Assembled into `DatabaseURL`     |
| `DB_PASSWORD` | postgres       |                                  |
| `DB_HOST`     | localhost      |                                  |
| `DB_PORT`     | 5432           |                                  |
| `DB_NAME`     | kitty_party    |                                  |

The final DSN is `postgres://user:pass@host:port/db?sslmode=disable`.

---

## 8. Cross-Cutting Concerns

| Concern          | Location                                                                |
| ---------------- | ----------------------------------------------------------------------- |
| Structured logs  | [internal/logger/logger.go](internal/logger/logger.go)                  |
| Request logging  | [internal/middleware/middleware.go](internal/middleware/middleware.go)  |
| Panic recovery   | ditto (`middleware.Recovery`)                                           |
| CORS             | ditto (`middleware.CORS`)                                               |
| JWT auth         | [internal/middleware/auth.go](internal/middleware/auth.go)              |
| Error types      | [internal/apperrors/errors.go](internal/apperrors/errors.go)            |
| Response helpers | [internal/response/response.go](internal/response/response.go)          |
| DB pool          | [internal/database/postgres.go](internal/database/postgres.go)          |

---

## 9. Testing

- Repository unit tests live next to their subjects: [member/repository_test.go](internal/member/repository_test.go), [group/repository_test.go](internal/group/repository_test.go), [group/membership_repository_test.go](internal/group/membership_repository_test.go).
- Run with `go test -race -v ./...` (or `make test`).
- Integration smoke tests against a running API are logged in [test_api.log](test_api.log) / [test_api.err.log](test_api.err.log).

---

## 10. Extending the System

To add a new bounded context `foo`:

1. Create [internal/foo/model.go](internal/foo/) with entities + request DTOs (use `binding` tags).
2. Define `Repository` interface in `repository.go`; implement it in `postgres_repository.go` with pgx.
3. Add a migration under [db/scripts/](db/scripts/) following the `V{n}__<name>.sql` pattern.
4. Define `Service` interface in `service.go`; enforce all business invariants here, not in the handler.
5. If you need data from another domain, define a narrow port (e.g., `FooNeedsBar`) and write an adapter that wraps `bar.Repository`.
6. Add `Handler` + `RegisterRoutes(rg *gin.RouterGroup)` in `handler.go`.
7. Add the handler to [router.Options](internal/router/router.go#L17) and wire the chain in [cmd/api/main.go](cmd/api/main.go#L54).

Do **not** import another domain's package from within a service — always go through an adapter/port. This is the invariant that keeps the dependency graph acyclic.
