// transaction_processor/processor/change_calculator.go
package processor

import (
	"fmt"

	"sig-takehome-exercise/models"
)

// ChangeCalculator is responsible for computing effective balance changes
type ChangeCalculator struct {
	snapshot *Snapshot
}

// NewChangeCalculator creates a new instance bound to a snapshot
func NewChangeCalculator(snapshot *Snapshot) *ChangeCalculator {
	return &ChangeCalculator{
		snapshot: snapshot,
	}
}

// CalculateChange returns the numeric value of a Change (either direct or referenced)
func (cc *ChangeCalculator) CalculateChange(change models.Change) (int64, error) {
	if change.IsDirectValue() {
		return change.GetDirectValue(), nil
	}

	if change.IsAccountReference() {
		balance, exists := cc.snapshot.Balances[change.Account]
		if !exists {
			balance = 0 // treat missing accounts as zero
		}

		switch change.Sign {
		case "plus":
			return int64(balance), nil
		case "minus":
			return -int64(balance), nil
		default:
			return 0, fmt.Errorf("invalid sign: %s", change.Sign)
		}
	}

	return 0, fmt.Errorf("change is neither direct value nor account reference")
}

// CalculateInstructionChanges resolves all instructions into net changes
func (cc *ChangeCalculator) CalculateInstructionChanges(instructions []models.Instruction) (map[string]int64, int64, error) {
	accountChanges := make(map[string]int64)
	total := int64(0)

	for i, inst := range instructions {
		val, err := cc.CalculateChange(inst.Change)
		if err != nil {
			return nil, 0, fmt.Errorf("instruction %d failed: %w", i, err)
		}
		accountChanges[inst.Account] += val
		total += val
	}

	return accountChanges, total, nil
}

// ValidateTransactionChanges ensures balance conservation and no negative balances
func (cc *ChangeCalculator) ValidateTransactionChanges(tx models.Transaction) error {
	accountChanges, total, err := cc.CalculateInstructionChanges(tx.Instructions)
	if err != nil {
		return fmt.Errorf("failed to compute changes: %w", err)
	}

	if total != 0 {
		return fmt.Errorf("sum of changes must be zero (got %d)", total)
	}

	for account, delta := range accountChanges {
		balance := cc.snapshot.Balances[account]
		if int64(balance)+delta < 0 {
			return fmt.Errorf("account %s would go negative: %d + (%d)", account, balance, delta)
		}
	}

	return nil
}
