# ADR-002: Money Representation

- Status: Accepted
- Date: 2026-09-04
- Confidence: High

## Context

The application needs to represent monetary amounts consistently across the domain,
database, repositories, and HTTP handlers.

Using floating-point types such as `float64` can introduce rounding errors and is
not suitable for storing financial values.

The application currently targets Indian Rupees (INR), where 1 Rupee is equal to
100 paise. All monetary values are represented using `int64`.

## Options Considered

### Option 1: `float64`

Represent amounts directly as decimal Rupee values.

### Option 2: `int64` in paise

Represent amounts as the smallest currency unit, where 1 INR = 100 paise.

### Option 3: Decimal type

Use a decimal/fixed-point library for monetary values.

## Decision

We use `int64` to represent monetary amounts in **paise (₹0.01)**.

For example:

- `₹1` = `100`
- `₹10.50` = `1050`
- `₹1000` = `100000`

All monetary values stored in the database and passed through the domain layer
are represented as integer paise. Floating-point types are not used for monetary
values.

## Tradeoffs

### Benefits

- No floating-point rounding errors.
- Exact representation of monetary values.
- Simple and efficient storage and arithmetic.
- Consistent representation across the application and database.
- `int64` provides a sufficiently large range for the application's expected
  monetary values.

### Costs

- Values are less human-readable without conversion to Rupees.
- The current representation is tied to paise as the smallest unit.
- Supporting currencies with different minor-unit conventions may require
  additional currency-specific handling.

## Consequences

All application layers must treat monetary `int64` values as paise.

Conversions to and from Rupees should happen only at system boundaries such as
HTTP request/response formatting or UI presentation.

For example, an API value of `1050` represents `₹10.50`.

Floating-point arithmetic must not be used for monetary calculations.

## Reconsideration

This decision should be revisited if:

- The application introduces multi-currency support.
- A currency with a different or variable minor-unit convention is supported.
- The application requires precision smaller than one paise.
- The range of `int64` becomes insufficient for supported monetary values.
