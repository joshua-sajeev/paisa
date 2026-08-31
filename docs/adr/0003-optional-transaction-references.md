# ADR-003: Nullable Foreign Keys for Transactions

* Status: Accepted
* Date: 2026-09-04
* Confidence: High

## Context

Transactions can represent different types of financial operations. Not every
transaction has both a source and destination account.

For example, an income may come from an external source and therefore has no
`from_account_id`. Similarly, an expense may go to an external party and
therefore has no `to_account_id`.

The database allows `from_account_id`, `to_account_id`, and `jar_id` to be
`NULL`, but the domain model must represent this possibility explicitly.

## Options Considered

### Option 1: Non-nullable Foreign Keys

Make all foreign keys `NOT NULL` and use a sentinel account to represent
external sources or destinations.

### Option 2: Nullable Foreign Keys

Allow foreign keys to be `NULL` when the relationship does not exist and
represent them as optional UUIDs in the domain model.

## Decision

We use **nullable foreign keys** for transaction relationships that are not
applicable to every transaction.

The domain model represents these fields using `*uuid.UUID`:

* `from_account_id`
* `to_account_id`
* `jar_id`

A `nil` value means that the transaction does not have that relationship.

Examples:

| Transaction      | `from_account_id` | `to_account_id` | `jar_id` |
| ---------------- | ----------------- | --------------- | -------- |
| Income           | `nil`             | Account         | `nil`    |
| Expense          | Account           | `nil`           | `nil`    |
| Account transfer | Account           | Account         | `nil`    |
| Jar transaction  | Account           | `nil`/Account   | Jar      |

The exact combination of fields is validated according to the transaction type.

## Tradeoffs

### Benefits

* Accurately represents real transaction relationships.
* No artificial "external" or sentinel accounts are required.
* Database constraints can still enforce valid referenced accounts when an ID
  is present.
* The domain model explicitly represents optional relationships.

### Costs

* Application code must handle `nil` UUIDs.
* Queries and joins must account for nullable foreign keys.
* Transaction validation becomes important to prevent invalid combinations.

## Consequences

All layers must treat these transaction relationships as optional.

The database uses nullable foreign keys, and the Go domain model uses
`*uuid.UUID` rather than `uuid.UUID` for these fields.

A `nil` foreign key has semantic meaning and must not be replaced with a
dummy UUID or sentinel record.

Transaction creation and update logic must validate that the combination of
fields is valid for the transaction type.

## Reconsideration

This decision should be revisited if:

* Transaction types change significantly.
* All transaction types eventually require both source and destination accounts.
* The application introduces external entities that should be modeled
  explicitly rather than represented by `NULL`.
