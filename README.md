# Paisa

Paisa is a personal finance tracker built with Go and PostgreSQL. It uses Hexagonal Architecture (Ports and Adapters) to maintain a clean separation between the core domain logic and infrastructure.

## Architecture

This project strictly adheres to **Hexagonal Architecture** (see [ADR-0001](docs/adr/0001-use-hexagonal-architecture.md)). 

- **Domain Core**: Business models and logic (independent of database, HTTP, or external systems).
- **Ports**: Interfaces defining how the core interacts with the external world (Inbound/Outbound).
- **Adapters**: Concrete implementations of ports (e.g., HTTP handlers, PostgreSQL repositories).

## Core Concepts

- **Accounts**: Track personal bank accounts, cards, or cash.
- **Jars**: Budgeting pools with customized allocation types and values (e.g., envelope system).
- **Goals**: Target savings goals with deadlines and automated contributions.
- **Transactions & Templates**: Record movement of money between accounts, assign them to Jars, track master income, and define templates for repeated patterns.

## Tech Stack

- **Backend**: Go (Golang)
- **Database**: PostgreSQL 18
- **Infras**: Docker & Docker Compose

## Getting Started

### Prerequisites

- Go (Golang)
- Docker and Docker Compose

## Render Deployment

The Render web service builds the frontend and backend from the repository root. Configure its build command as:

```sh
cd frontend && npm ci && npm run build && cd ../backend && go build -o paisa .
```

`npm ci` installs the frontend dependencies from `frontend/package-lock.json` before the Next.js build. When using a custom Render build command, include this install step; otherwise `next build` can fail with `next: not found`.

## License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.
