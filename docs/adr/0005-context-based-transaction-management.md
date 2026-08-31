# ADR-0005: Context-Based Transaction Management (`TxManager`)

- Status: Accepted
- Date: 2026-09-05
- Confidence: High

## Context

As I build out the Paisa application, features like creating financial transactions with associated jar allocations require multi-operation atomicity (all steps must succeed or roll back together). 

I faced a structural challenge: how to manage database transactions across multiple repositories within the service layer without leaking database-specific primitives (such as `pgx.Tx`) into pure domain port interfaces (`ports.AccountRepository`, etc.).

## Options Considered

### Option 1: Explicit Transaction Passing

Pass the active database transaction (`pgx.Tx`) directly as an argument into every repository method signature.

* Change methods from `Create(ctx, data)` to `Create(ctx, tx, data)`.
* **Drawback:** It pollutes domain ports and use cases with infrastructure-specific driver details, tightly coupling the application layer to PostgreSQL.

### Option 2: Context-Based Transaction Management (`TxManager`)

Manage transactions via a wrapper and inject the active transaction into the `context.Context`.

* A service-level `TxManager` handles transaction lifecycles (`Begin`, `Commit`, `Rollback`) and stores the transaction inside the `context.Context`. Repositories transparently extract it using an internal helper (`dbExecutor`).

### Option 3: No Transaction Manager

Rely on independent queries or handle rollbacks manually inside services.

* Skip transaction orchestration layers.
* **Drawback:** Leaves multi-step workflows vulnerable to data corruption and partial writes during failures.

## Decision

I will implement **Option 2: Context-Based Transaction Management (`TxManager`)**. 

The service layer will control transaction boundaries using a `TxManager` interface, while all repository implementations will use an internal helper (`dbExecutor`) to transparently resolve either the active transaction from the context or fall back to the connection pool. Port interfaces will remain completely clean of infrastructure concerns.

## Tradeoffs

### Benefits

- **Decoupled Architecture:** Domain ports and use cases remain pure and agnostic of database driver types.
- **Uniformity:** All repositories (accounts, transactions, jars, etc.) can be wired up uniformly to support transactions from day one.
- **Developer Ergonomics:** Service methods only deal with standard `context.Context` when orchestrating multi-step atomic writes.

### Costs

- **Boilerplate:** Requires a small internal `dbExecutor` helper function in each repository implementation.
- **Context Value Usage:** Relies on `context.Value` lookup keys, which are opaque and hidden compared to explicit parameters.

## Consequences

- All current and future persistence port implementations in package `postgres` must adopt the `dbExecutor` pattern.
- Multi-repo or multi-step operations can now safely wrap business logic in `TxManager.WithinTransaction` blocks with guaranteed atomicity.

## Reconsideration

This decision should be revisited if:
- Go introduces standard, language-level context transaction propagation.
- The project migrates to a heavy ORM framework that provides a built-in Unit of Work implementation.
