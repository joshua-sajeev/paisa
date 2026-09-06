# ADR-0008: Login Rate Limiting

- Status: Accepted
- Date: 2026-09-06
- Confidence: High

## Context

PIN-based authentication (ADR-0006) is susceptible to brute-force attacks. An attacker with network access could attempt many PIN combinations without throttling. For a 6 digit PIN, this represents a limited keyspace (1,000,000 possibilities). I need to prevent rapid-fire login attempts while maintaining usability for legitimate users who occasionally mistype their PIN.

## Options Considered

### Option 1: Per-IP Rate Limiting (Selected)

Rate limit login attempts by client IP address using the token bucket algorithm. Each IP gets an independent limit (e.g., 5 attempts per minute, burst of 1).

**Pros**:
- Simple to implement
- Effective against brute-force attacks
- Per-client isolation (one attacker doesn't affect others)
- Uses Go standard library (`golang.org/x/time/rate`)
- Legitimate users rarely hit the limit

**Cons**:
- Doesn't work behind load balancers without `X-Forwarded-For` parsing
- Mobile users switching networks see different IPs
- Shared networks (corporate, school) may block legitimate users
- No persistent state across restarts (attackers resume after app restart)

### Option 2: Global Rate Limiting

Single global limit on all login attempts across all IPs (e.g., 50 attempts per minute total).

**Pros**:
- Simple implementation
- Protects against distributed attacks

**Cons**:
- One attacker can deny service to all legitimate users
- Not suitable for single-user app
- Too coarse-grained

### Option 3: Account Lockout

Lock the account temporarily after N failed attempts (e.g., lock after 5 failures for 15 minutes).

**Pros**:
- Strong brute-force protection
- User-aware (knows they're locked out)

**Cons**:
- Adds account state complexity
- Single-user app has no account recovery flow
- Worse UX for user who forgets PIN

### Option 4: No Rate Limiting

Accept brute-force risk; rely on PIN entropy and local device protection.

**Pros**:
- Simplest implementation

**Cons**:
- Vulnerable to brute-force
- No protection for weak PINs

## Decision

I chose Option 1: Per-IP rate limiting. It provides effective brute-force protection without adding complex account state. The token bucket algorithm allows occasional mistakes while blocking sustained attacks.

**Configuration**:
- Rate: 5 attempts per minute sustained
- Burst: 1 (no allowance for quick retries)

## Tradeoffs

### Benefits

- Brute-force resistant: Attacker cannot quickly exhaust PIN keyspace
- Stateless: No database or persistent storage needed
- Development-friendly: In-memory limiter works with current session store
- Fair: Each IP has independent limit
- Standard algorithm: Token bucket is industry-standard

### Costs

- Network address dependency: Requires accurate client IP extraction
- Temporary blocking: Legitimate user mistakes trigger brief lockout
- No persistence: Limiter state lost on restart (minor risk)

## Consequences

- **Reduced attack surface**: Brute-force attacks now require hours instead of seconds
- **Operational**: Must ensure `clientIP()` extraction handles proxies correctly (watch for `X-Forwarded-For` in production)
- **UX**: User hitting rate limit receives HTTP 429 with clear message
- **Logging**: Rate limit hits are logged at WARN level for monitoring
- **Scalability**: In-memory limiter holds one entry per unique IP (bounded by concurrent attackers)

## Reconsideration

Revisit this decision if:
- I need to support account lockout or persistent rate limiting across restarts
- The app runs behind a proxy that obscures client IP
- Multi-user support is added (may need per-user instead of per-IP limiting)
- PIN entropy is reduced below current standard (6 digits)

---

**Implementation**: `backend/internal/adapter/http/login_rate_limiter.go`
