# ADR-002: Money Representation

* Status: Accepted
* Date: 2026-09-04
* Confidence: High

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

* `₹1` = `100`
* `₹10.50` = `1050`
* `₹1000` = `100000`

All monetary values stored in the database and passed through the domain layer
are represented as integer paise. Floating-point types are not used for
monetary values.

### Non-monetary `int64` values

Not every `int64` in the domain represents money.

For example, `Jar.AllocationValue` represents an **allocation percentage**, not
a monetary amount:

* Percentage jar: `1–100`
* Remainder jar: `0`

Therefore, the paise convention applies only to fields that represent monetary
amounts. Domain models must document the unit and semantics of other numeric
fields explicitly.

## Tradeoffs

### Benefits

* No floating-point rounding errors.
* Exact representation of monetary values.
* Simple and efficient storage and arithmetic.
* Consistent representation across the application and database.
* `int64` provides a sufficiently large range for the application's expected
  monetary values.
* Numeric fields with different semantics can remain simple `int64` values when
  their units are explicitly documented.

### Costs

* Monetary values are less human-readable without conversion to Rupees.
* The current monetary representation is tied to paise as the smallest unit.
* Supporting currencies with different minor-unit conventions may require
  additional currency-specific handling.
* Developers must distinguish monetary `int64` values from other numeric
  `int64` values such as percentages, counts, and identifiers.

## Consequences

All application layers must treat **monetary** `int64` values as paise.

Conversions to and from Rupees should happen only at system boundaries such as
HTTP request/response formatting or UI presentation.

For example, an API value of `1050` represents `₹10.50`.

Floating-point arithmetic must not be used for monetary calculations.

Other numeric fields must document their own unit and valid range where their
meaning is not self-evident.

## Reconsideration

This decision should be revisited if:

* The application introduces multi-currency support.
* A currency with a different or variable minor-unit convention is supported.
* The application requires precision smaller than one paise.
* The range of `int64` becomes insufficient for supported monetary values.
* Monetary calculations require capabilities that cannot be safely or
  conveniently implemented using integer paise.
