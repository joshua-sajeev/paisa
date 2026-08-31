# ADR-0001: Use Hexagonal Architecture

* Status: Accepted
* Date: 2026-08-29
* Confidence: High

## Context

The project needs an architecture that keeps business logic independent from infrastructure such as HTTP, PostgreSQL, and external services.

## Options Considered

### Option 1: Hexagonal Architecture

Separates the application core from infrastructure using ports and adapters.

### Option 2: Layered Architecture

Separates the application into layers such as handlers, services, and repositories.

### Option 3: Clean Architecture

Provides strict separation between domain, application, and infrastructure layers.

## Decision

Use **Hexagonal Architecture** to keep the core independent of external dependencies and make the system easier to test and maintain.

## Tradeoffs

### Benefits

* Clear separation of concerns.
* Easier testing.
* Infrastructure can be replaced independently.

### Costs

* More interfaces and abstractions.
* Slightly more initial complexity.

## Consequences

New external dependencies should be integrated through adapters and ports rather than directly into the application core.

## Reconsideration

Revisit this decision if the architecture introduces unnecessary complexity or no longer fits the project's requirements.
