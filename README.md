# Sig Take-Home Exercise

This repository contains Part 2 of the take-home assignment for SIG. It implements a simplified Solana-like transaction validator with real-time batching, conflict handling, state snapshots, and UDP/HTTP communication between client and server.

---

## Part 1: Leader Phase Design

The architectural analysis and breakdown of the `leader` phase from the Solana validator is provided in:

**`leader-phase/leader-phase.md`**

This file covers the responsibilities and internal flow of the Banking Stage, Scheduler, PoH integration, transaction execution, and commitment pipeline based on Solana's Agave implementation.

---

## Part 2: Validator Implementation

### Design Overview

This project simulates a simplified validator node that:

* Listens for transactions on **UDP port 2001**.
* Batches incoming transactions up to 100 at a time.
* Posts each batch to an HTTP server (simulating downstream consumers).
* Applies valid transactions locally, including fee deduction and balance mutation.
* Takes a snapshot of account balances after each batch.
* Logs real-time statistics like total transactions, fee accumulation, and invalid counts.

### Key Components

* `main.go` – Entry point, bootstraps handler, processor, and batch sender.
* `processor/` – Core logic for transaction validation, application, and batching.
* `models/transactions.go` – Transaction structure and input parsing.
* `test/send_tx.go` – Sends test transactions to UDP port 2001.
* `test_server/server.go` – Receives HTTP batches and prints them (runs on port 2002).
* `scripts/` – Build and run scripts.
* `Makefile` – Easy automation of tasks like build, run, test, and reset.

---

## How to Run

### Prerequisites

* Go 1.20+
* Linux or macOS with terminal support
* `make`, `lsof`, and `kill` utilities (standard on Unix)

* Install Go:

  - **macOS**: Use Homebrew
    ```bash
    brew install go
    ```

  - **Linux (Debian/Ubuntu)**:
    ```bash
    sudo apt update
    sudo apt install golang-go
    ```

### Step 1: Clone and Setup

```bash
$ git clone https://github.com/naman20sharma/sig-takehome-exercise.git
$ cd sig-takehome-exercise
$ go mod tidy
```

Make sure `go.mod` is present in the root directory.

### Step 2: Provide Snapshot

Place your input account snapshot in the root directory with the filename:

```bash
accounts.json
```

This file contains initial balances and is used when starting the validator.

### Step 3: Run everything together (auto-detects OS)

```bash
$ make full
```

This builds the validator, starts the HTTP server in one terminal and the validator in another.

### Step 4: Send test transactions

```bash
$ make test
```

This sends a batch of test transactions from `test/send_tx.go` to the validator over UDP.

### Optional: Run manually if `make full` doesn’t launch terminals correctly

```bash
$ make build       # Build the validator binary
$ make server      # Run HTTP server (on :2002)
$ make run         # Start validator (UDP :2001)
```

### Optional: Reset environment

To stop all processes and start fresh:

```bash
$ make reset
$ make clean
```

This kills any existing validator/server, frees UDP port 2001 and HTTP port 2002, and deletes the `bin/validator` binary.

---

## File I/O Notes

* The validator expects `accounts.json` as input.
* After each batch, a new snapshot is written to the root directory as:

```
accounts-T-{timestamp}.json
```

These reflect updated balances.

---

## Transaction Format

Each transaction has the following JSON schema:

```json
{
  "fee": { "payer": "alice", "amount": 3 },
  "instructions": [
    { "account": "bob", "change": 5 },
    { "account": "charlie", "change": { "account": "alice", "sign": "minus" } }
  ]
}
```

---

## Design Decisions

* **Conflict Resolution**: To avoid double-spending or race conditions, each batch avoids including multiple transactions that touch the same accounts. Conflicting ones are deferred to future batches.

* **Batching**: Max 100 transactions per batch, max 100 batches/sec. This simulates real-time leader tick batching.

* **Loose Coupling**: The UDP listener, batch sender, and snapshot saver are fully decoupled and communicate via in-memory buffers.

* **Logging**: Transaction stats (received, processed, fees, etc.) are logged every 30s.

* **Cross-Platform Support**: `make full` detects `Darwin` (macOS) or Linux and launches new terminals accordingly.

---

## Trade-offs

* **Error Tolerance**: Invalid transactions are logged but do not interrupt the processing of the batch. This increases fault tolerance but risks overlooking frequent or critical transaction issues if not monitored carefully.

* **No Persistent Queueing**: Transactions are held in memory before batching. If the validator crashes before processing them, the in-flight transactions are lost. A persistent queue (e.g., file-based or Redis) could improve fault tolerance but adds complexity.

* **Simple Conflict Handling**: Transactions that conflict are deferred, not re-ordered or partially applied. This simplifies logic but can increase latency for high-volume account interactions.

* **Fixed Port Binding**: The server and validator use hardcoded ports (2001 for UDP, 2002 for HTTP). While straightforward, this limits flexibility in multi-instance environments unless recompiled or updated.

* **No TLS or Auth**: The UDP and HTTP endpoints are unencrypted and unauthenticated, making them unsuitable for production without wrapping layers or reverse proxies.

---

## Submission

Please ensure both parts are present:

* Part 1: `leader-phase/leader-phase.md`
* Part 2: This project and source code

Thank you for the opportunity!
