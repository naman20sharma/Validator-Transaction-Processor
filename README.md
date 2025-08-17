# Validator Transaction Processor

A simplified Solana-inspired transaction validator written in Go.  
It demonstrates real-time batching, conflict handling, state snapshots, and UDP/HTTP communication between client and server. This project is meant as an educational exploration of validator design and can serve as a base for experimenting with distributed systems and blockchain-style transaction flow.

## Background

The design draws inspiration from Solana’s Agave validator. Alongside the code, I have written a breakdown of the *leader phase* — covering responsibilities such as Banking Stage, Scheduler, Proof-of-History (PoH) integration, transaction execution, and commitment pipeline.

See: **`leader-phase/leader-phase.md`**


## Features

- Listens for transactions on **UDP port 2001**  
- Batches incoming transactions (up to 100 at a time)  
- Posts each batch to an HTTP server (simulating downstream consumers)  
- Applies valid transactions locally with fee deduction and balance mutation  
- Generates snapshots of account balances after each batch  
- Logs real-time statistics such as transaction counts, fee accumulation, and invalid counts  

## Project Structure

- `main.go` – Entry point, bootstraps handler, processor, and batch sender  
- `processor/` – Core logic for validation, application, and batching  
- `models/transactions.go` – Transaction structure and input parsing  
- `test/send_tx.go` – Sends test transactions to UDP port 2001  
- `test_server/server.go` – Simple HTTP batch receiver (port 2002)  
- `scripts/` – Build and run scripts  
- `Makefile` – Common automation for build, run, test, reset  

## Getting Started

### Prerequisites
- Go 1.20+  
- Linux or macOS  
- `make`, `lsof`, and `kill` utilities  

### Install Go

Install Go (example for macOS with Homebrew):
```bash
brew install go
```

### Clone and Setup
```bash
git clone https://github.com/<your-username>/validator-transaction-processor.git
cd validator-transaction-processor
go mod tidy
```

### Provide Initial Snapshot

Place your starting account balances in the root as accounts.json.

### Run Everything
```bash
make full
```

This builds the validator, starts the HTTP server, and launches the validator.

### Send Test Transactions
```bash
make test
```

### Reset Environment
```bash
make reset
make clean
```

## Transaction Format

Example JSON transaction:

```json
{
  "fee": { "payer": "alice", "amount": 3 },
  "instructions": [
    { "account": "bob", "change": 5 },
    { "account": "charlie", "change": { "account": "alice", "sign": "minus" } }
  ]
}
```

## Design Decisions

- Conflict Resolution: Transactions that touch the same accounts are deferred to future batches.
- Batching: Max 100 transactions per batch, max 100 batches/sec.
- Loose Coupling: UDP listener, batch sender, and snapshot saver are decoupled with in-memory buffers.
- Logging: Stats logged every 30s.
- Cross-Platform Support: make full adapts for macOS and Linux.

## Trade-offs and Future Improvements

- Transactions are held in memory, so a crash drops in-flight transactions. A persistent queue would help.
- Conflict handling is simple (defer only) which may increase latency under contention.
- Ports are hardcoded (2001/2002). Configurable ports would be more flexible.
- No TLS or authentication is provided — suitable for experiments, not production.
