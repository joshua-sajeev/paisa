# ADR-0007: Session Management

- Status: Accepted
- Date: 2026-09-06
- Confidence: High

## Context

PIN authentication (ADR-0006) is stateless. I need to maintain user sessions across HTTP requests, track session state, enforce TTL, support logout, and enable demo mode. Session management affects security, performance, and scalability. Currently targeting single-user, single-instance deployment.

## Options Considered

### Option 1: In-Memory Session Store (Selected)

Sessions stored in application memory (map of session ID → session data).

**Pros**:
- Fast O(1) lookup
- No external dependencies
- Simple implementation
- Suitable for single-instance deployment

**Cons**:
- Sessions lost on restart
- Cannot scale horizontally
- Memory leak risk if cleanup fails
- No persistence for audit trails

### Option 2: Redis / Distributed Cache

Sessions persisted in Redis.

**Pros**:
- Persists across restarts
- Horizontal scaling support
- Industry standard
- Native TTL support

**Cons**:
- Extra infrastructure dependency
- Network latency per session lookup
- Operational complexity
- Overkill for single-user app

### Option 3: Database-Backed Sessions

Sessions in PostgreSQL table.

**Pros**:
- Single database for all data
- Audit trail available
- Survives restarts

**Cons**:
- Database latency per lookup
- Requires schema migrations
- Complex cleanup of expired sessions

### Option 4: Stateless Token (JWT)

Signed JWT with embedded user claims, no server-side storage.

**Pros**:
- Truly stateless
- Unlimited horizontal scaling
- Fast validation

**Cons**:
- Cannot revoke immediately on logout
- Requires token blacklist for logout
- Tokens contain user data
- Complex token refresh logic

## Decision

I chose Option 1: In-memory session store. Fast, simple, no external dependencies—perfect for single-user, single-instance app. Migration path to Redis/database exists if deployment changes.

## Tradeoffs

### Benefits

- Simple implementation: In-memory map
- Fast lookup: O(1) per request
- No infrastructure: Single app instance
- Developer friendly: Easy to debug/mock
- Suitable for current scale

### Costs

- Sessions lost on restart: User re-logs in with PIN
- No horizontal scaling: Multiple instances can't share sessions
- Memory management: Must clean expired sessions
- No audit trail: Cannot query historical sessions
- Single-machine deployment only

## Consequences

- Session ID generated after PIN auth
- Session stored with UserID, CreatedAt, ExpiresAt
- TTL checked on retrieval (default 24 hours, configurable)
- Background cleanup removes expired sessions
- Logout calls invalidation handler
- Bearer token transport in `Authorization` header

## Reconsideration

Revisit if:
- High availability required (migrate to Redis)
- Horizontal scaling needed (migrate to Redis)
- Audit trail of all sessions required (migrate to database)
- Session persistence across restarts required

---

**Implementation**: `backend/internal/session/`, `backend/internal/adapter/http/auth_middleware.go`
