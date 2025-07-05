package processor

import (
	"log"
	"sync"

	"sig-takehome-exercise/models"
)

type TransactionProcessor struct {
	snapshot      *Snapshot
	validatorAcct string
	calculator    *ChangeCalculator
	balanceMutex  sync.Mutex
}

// NewTransactionProcessor initializes the processor with loaded snapshot
func NewTransactionProcessor(snapshot *Snapshot) *TransactionProcessor {
	return &TransactionProcessor{
		snapshot:      snapshot,
		validatorAcct: "validator",
		calculator:    NewChangeCalculator(snapshot),
	}
}

// ValidateTransaction ensures a transaction is well-formed and won’t violate balances
func (p *TransactionProcessor) ValidateTransaction(tx models.Transaction) bool {
	p.balanceMutex.Lock()
	defer p.balanceMutex.Unlock()

	// Ensure fee payer has enough balance
	feePayerBalance := p.snapshot.Balances[tx.Fee.Payer]
	if feePayerBalance < tx.Fee.Amount {
		return false
	}

	// Use ChangeCalculator to verify instruction validity
	if err := p.calculator.ValidateTransactionChanges(tx); err != nil {
		return false
	}

	return true
}

// ApplyTransaction updates balances by applying fee always,
// then applies instructions only if they pass validation.
func (p *TransactionProcessor) ApplyTransaction(tx models.Transaction) error {
	p.balanceMutex.Lock()
	defer p.balanceMutex.Unlock()

	// 1) Deduct fee from payer and credit validator
	p.snapshot.Balances[tx.Fee.Payer] -= tx.Fee.Amount
	p.snapshot.Balances[p.validatorAcct] += tx.Fee.Amount

	log.Printf("\033[32m Fee of %d deducted from %s and credited to validator\033[0m",
		tx.Fee.Amount, tx.Fee.Payer)
	// 2) Try to validate instruction changes
	if err := p.calculator.ValidateTransactionChanges(tx); err != nil {
		log.Printf("\033[33m Skipping instructions for tx from %s: %v\033[0m",
			tx.Fee.Payer, err)
		// instructions fail → fee applied, instructions skipped
		return nil
	}

	// 3) If valid, compute and apply instruction changes
	accountChanges, _, _ := p.calculator.CalculateInstructionChanges(tx.Instructions)
	for acct, delta := range accountChanges {
		p.snapshot.Balances[acct] += delta
	}
	log.Printf("\033[32m Instructions applied successfully for tx from %s\033[0m", tx.Fee.Payer)
	return nil
}

// GetSnapshot returns the current snapshot (used by batch or saving modules)
func (p *TransactionProcessor) GetSnapshot() *Snapshot {
	return p.snapshot
}
