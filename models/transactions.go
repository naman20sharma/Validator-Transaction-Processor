package models

import (
	"encoding/json"
	"fmt"
)

type Fee struct {
	Payer  string `json:"payer"`
	Amount int64  `json:"amount"`
}

type Change struct {
	Account string `json:"account,omitempty"`
	Sign    string `json:"sign,omitempty"`
	Value   *int64 // used if it's a fixed number
}

// Custom unmarshaler to support int or object for `change`
func (c *Change) UnmarshalJSON(data []byte) error {
	// Try to unmarshal as a number first
	var num int64
	if err := json.Unmarshal(data, &num); err == nil {
		c.Value = &num
		c.Account = ""
		c.Sign = ""
		return nil
	}

	// Try to unmarshal as float64 (JSON default for numbers)
	var floatNum float64
	if err := json.Unmarshal(data, &floatNum); err == nil {
		intVal := int64(floatNum)
		c.Value = &intVal
		c.Account = ""
		c.Sign = ""
		return nil
	}

	// Try to unmarshal as object
	var obj struct {
		Account string `json:"account"`
		Sign    string `json:"sign"`
	}
	if err := json.Unmarshal(data, &obj); err != nil {
		return fmt.Errorf("change must be either a number or an object with 'account' and 'sign': %w", err)
	}

	// Validate the sign field
	if obj.Sign != "plus" && obj.Sign != "minus" {
		return fmt.Errorf("sign must be 'plus' or 'minus', got: %s", obj.Sign)
	}

	c.Account = obj.Account
	c.Sign = obj.Sign
	c.Value = nil
	return nil
}

// IsDirectValue returns true if this is a direct numeric value
func (c *Change) IsDirectValue() bool {
	return c.Value != nil
}

// IsAccountReference returns true if this references another account
func (c *Change) IsAccountReference() bool {
	return c.Account != ""
}

// GetDirectValue returns the direct numeric value (panics if not a direct value)
func (c *Change) GetDirectValue() int64 {
	if c.Value == nil {
		panic("attempting to get direct value from account reference")
	}
	return *c.Value
}

type Instruction struct {
	Account string `json:"account"`
	Change  Change `json:"change"`
}

type Transaction struct {
	Fee          Fee           `json:"fee"`
	Instructions []Instruction `json:"instructions"`
}

// ValidateBasicStructure performs basic validation on the transaction structure
func (t *Transaction) ValidateBasicStructure() error {
	if t.Fee.Payer == "" {
		return fmt.Errorf("fee payer cannot be empty")
	}
	if t.Fee.Amount <= 0 {
		return fmt.Errorf("fee amount must be positive, got: %d", t.Fee.Amount)
	}
	if len(t.Instructions) == 0 {
		return fmt.Errorf("transaction must have at least one instruction")
	}
	for i, instruction := range t.Instructions {
		if instruction.Account == "" {
			return fmt.Errorf("instruction %d: account cannot be empty", i)
		}
	}
	return nil
}
