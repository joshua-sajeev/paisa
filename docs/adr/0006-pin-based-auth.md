# ADR-0006: PIN-Based Authentication

- Status: Accepted
- Date: 2026-09-06
- Confidence: High

## Context

Paisa is a single-user personal finance app requiring authentication for sensitive data. Traditional username/password adds friction and maintenance burden. PIN-based auth is standard in financial services (banking apps, payment systems) and matches my threat model: local device protection, not distributed multi-user access control.

## Options Considered

### Option 1: PIN-Based Authentication (Selected)

Numeric PIN (6 digits) set during setup, hashed with Argon2.

**Pros**:
- Simpler mental model
- Standard in financial apps
- Faster onboarding
- Matches single-user threat model

**Cons**:
- Limited entropy vs passwords
- Single-user assumption
- Requires robust hashing (Argon2)
- No traditional password recovery

### Option 2: Username + Password

Standard account creation with strong password requirements.

**Pros**:
- Familiar pattern
- Industry best practices available

**Cons**:
- Higher onboarding friction
- Password management burden
- Password reset infrastructure needed
- Overkill for single-user app

### Option 3: OAuth / Social Login

Delegate to Google, GitHub, etc.

**Pros**:
- No password management
- Industry-standard security

**Cons**:
- External dependencies
- Privacy concerns with financial data
- Misaligned with personal finance philosophy

### Option 4: Magic Link / Passwordless

Email or SMS link for login.

**Pros**:
- Modern UX
- No password storage

**Cons**:
- Email/SMS infrastructure needed
- Delivery delays
- Complex token handling
- Not suitable for offline-first app

## Decision

I chose Option 1: PIN-based authentication. It provides fast setup/login with acceptable security for my threat model, eliminates password recovery complexity, and aligns with financial app conventions.

**PIN standard**: 6 digits, Argon2 hashing.

## Tradeoffs

### Benefits

- Simpler onboarding: Set PIN, use immediately
- No password recovery: Eliminates support burden
- Faster login: 6 digits faster than username+password
- Aligned expectations: Users expect PINs in financial contexts
- Lower complexity: Fewer subsystems

### Costs

- Lower entropy: 6 digits < password complexity
- Single-user only: No multi-user per device without redesign
- PIN compromise: Requires account recovery mechanism (not traditional password reset)
- No username: Can't switch users on same device

## Consequences

- Session generation after successful PIN auth (see ADR-0007)
- PIN reset requires secure mechanism (future consideration)
- Demo mode allows auth bypass for testing
- Rate limiting on PIN attempts (see ADR-0008)

## Reconsideration

Revisit if:
- Multi-user support becomes required
- Higher entropy is needed (migrate to password-based auth)
- Offline authentication becomes critical

---

**Implementation**: `backend/internal/security/pin.go`, `backend/internal/application/auth_service.go`
