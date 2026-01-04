# Expense Tracker (Hexagonal Architecture)
A secure, multi-user financial tracking application built with a focus on clean architecture and high testability.
This project demonstrates a Hexagonal (Ports and Adapters) structure in Go, coupled with a SvelteKit frontend.
## Project Overview
The Expense Tracker allows users to manage transactions, track budgets, keep an eye on wallets and on depots (including stocks).
It is designed to be lightweight enough for personal use but architecturally robust enough to scale or switch infrastructure (like databases) with minimal effort.
### Core Design Goals
 + **Decoupling**: Business logic (Core) has zero dependencies on external frameworks or databases.
 + **Portability**: The system can run in "Demo Mode" (In-Memory) or "Production Mode" (Postgres) via environment configuration.
 + **Security** Multi-user isolation enforced at the service layer; passwords managed via Bcrypt hashing.
## Tech Stack
 + **Backend**: Go 1.24 (Standard Library for routing, bcrypt for security).
 + **Frontend** SvelteKit (Single-page CRUD interface).
 + **Database** Neon (Serverless Postgres) for production; Thread-safe Maps for local/demo.
 + **Infrastructure** Docker, Render (Hosting), Neon (Postgresql DB), GitHub+Actions (CI/CD).
 + **Security** Cookie-based session management with HttpOnly/Secure/SameSite flags.
## Project Structure
```
├── backend/
│   ├── cmd/server/         # Composition Root: Wires dependencies, routes and starts the server
│   ├── pkg/auth/           # Common Utils (hashing,..)
│   ├── scripts/            # Common scripts (sql schema for initial db-setup, manual password hash,..)
│   ├── internal/core/
│   │   ├── domain/         # Pure Business Objects & Domain Errors
│   │   ├── ports/          # Interfaces defining Driving (API) and Driven (DB) contracts
│   │   └── services/       # Implementation of business rules (The "Core")
│   └── adapters/
│       ├── handler/
│       │   ├── http/       # Driving Adapter: Translates HTTP to Core calls
│       │   └── middleware/ # Middleware to wrap calls to /api/* with authentication
│       └── repository/     # Driven Adapters: Postgres and In-Memory implementations
├── frontend/               # SvelteKit application (Static assets embedded in Go)
├── assets.go               # Go embed configuration for frontend assets
└── Dockerfile              # Dockerfile (▀̿Ĺ̯▀̿ ̿)
```
## Setup & Installation
### Prerequisites
 + Go 1.24+
 + (Optional) Docker
 + (Optional) A Postgres connection string
### Local Development (In-Memory)
Run the application without a database:
```
export APP_ENV=demo
go run backend/cmd/server/main.go
```
Alternatively, if neither `APP_ENV` nor `DATABASE_URL` are set, it will default to demo-mode.
The app will be available at `http://localhost:8080`.
### Production Setup
1. Set the `DATABASE_URL` environment variable to your Postgres string.
2. Set `APP_ENV=production` (or leave blank; if `DATABASE_URL` is set, the default is production mode).
3. Initialize the schema using `scripts/schema.sql` in your Database.
4. Manually insert User into DB (you might want to use `scripts/create_password_hash.go`).
5. Run `go run backend/cmd/server/main.go`.
## Architecture & Design Notes
### Dependency Injection
Dependencies are injected at the Composition Root (`main.go`).
Services receive a `ports.ExpenseRepository` interface.
This allows us to swap between `postgres.Repository` and `memory.Repository` without changing a single line of business logic.
### API Design
Login/Logout is handled via the `/auth`-Route.
Other than that, the API follows a resource-oriented RESTful structure under the `/api/` prefix (e.g. `GET` on `/api/transactions` returns all transaction tied to the active user if there is a valid session).
 * **Separation of Concerns**: Handlers (adapters/handler/http) are responsible for parsing JSON and managing HTTP status codes, while the Service layer handles ownership checks and business rules.
 * **Stateless Communication**: Aside from the session cookie, the API remains stateless. Each request carries the user context required to perform isolated operations.
 * **Response Pattern**: Successful mutations return 201 Created or 204 No Content. Error handling is unified; domain-specific errors (e.g., ErrBudgetNotFound) are mapped to appropriate HTTP status codes (e.g., 400 Bad Request) at the adapter level.
 There shouldn't be any surprises here.
### Authentication & Middleware
Security is handled via a non-intrusive middleware:
 1. **Login**: Verifies credentials against Bcrypt hashes stored in the DB.
 2. **Session** Sets a `session_token` cookie with `SameSite=strict`.
 3. **Guard** `AuthMiddleware` intercepts protected requests, extracts the `UserID` from the cookie, and injects it into the request Context.
 4. **Context** Services pull the `UserID` from the context to ensure a user can only view/edit their own data.
### "Demo Mode" Strategy
To facilitate testing and showcases while keeping my own instance encapsulated, there is a Demo-Mode.
Demo-Mode is set via `.env`-Variable and is therefore separated from production instance.
In Demo-Mode
 * the repository falls back to the in-memory implementation (persistence not needed).
 * DemoAuthMiddleware is injected, bypassing authentication
 * data can be read by everyone and will be deleted eventually!
## Testing
Unit tests are located alongside the services in `backend/internal/core/services/`.
The goal is to thoroughly unit-test the core and have a few happy-path integration-tests.
### Running Tests
`go test ./...`
The tests utilize the `memory` repository as a natural mock, ensuring fast execution and reliable verification of business rules.
## Limitations & Known Trade-offs
 + **Manual User Creation**: There is no public sign-up endpoint.
 Users must be added to the database manually (Hashed password + Username) to maintain a closed, secure environment.
## Future Improvements
 - [ ] Add a "Search" port to enable full-text search on transaction descriptions.
 - [ ] Integrate a SvelteKit-based dashboard for financial visualizations.
## License
Refer to the LICENSE file in the root directory.