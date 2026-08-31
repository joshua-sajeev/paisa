# ADR-004: Jar Allocation Strategy

* Status: Accepted
* Date: 2026-09-04
* Confidence: High

## Context

Paisa supports allocating a master income across multiple jars.

Each jar allocation can be configured as either:

* `percentage` — allocates a specified percentage of the master income.
* `remainder` — receives the amount remaining after all percentage allocations.

A master income represents the original income amount. Jar allocations represent
how that income is distributed among the configured jars.

The `JarAllocationService` manages allocation configuration, while the allocation
calculation is domain logic that determines the actual amount assigned to each jar.

## Options Considered

### Option 1: Manual Allocation

Allow the caller to specify the amount allocated to each jar for every master
income transaction.

### Option 2: Percentage and Remainder Allocation

Calculate jar amounts automatically from the configured allocation percentages,
with one remainder jar receiving the remaining amount.

## Decision

We use **percentage and remainder based allocation** for master income
transactions.

For a master income of `M`:

1. Each percentage allocation receives its configured percentage of `M`.
2. The remainder jar receives the amount left after all percentage allocations.
3. Only one remainder allocation is allowed.
4. The total allocated amount must equal the master income.

For example, for a master income of `₹10,000`:

* Needs jar: 50% → `₹5,000`
* Savings jar: 30% → `₹3,000`
* Spending jar: remainder → `₹2,000`

Allocation values are stored as integers. The remainder allocation does not
require an allocation value and is stored as `NULL`.

## Tradeoffs

### Benefits

* Automatically distributes income according to the user's jar configuration.
* Prevents the allocated amounts from exceeding or falling short of the master
  income.
* A remainder jar handles amounts that cannot be evenly represented due to
  integer paise rounding.
* Users do not need to manually calculate individual jar amounts.

### Costs

* Allocation configuration requires validation.
* Only one remainder allocation can exist.
* Changes to allocation configuration can affect future master income
  allocations.

## Consequences

The allocation engine is responsible for calculating the actual amount assigned
to each jar when a master income is processed.

Percentage allocations are calculated from the master income amount, which is
represented in paise.

The remainder allocation receives the difference between the master income and
the total of all percentage allocations.

Once allocations for a master income have been created, they are treated as
immutable. Changes to jar allocation configuration apply only to future master
income transactions.

## Reconsideration

This decision should be revisited if the application needs:

* Fixed-amount jar allocations.
* Multiple remainder jars.
* Custom allocation rules.
* Editable historical allocations.
* Allocation rules that depend on transaction categories or other conditions.
