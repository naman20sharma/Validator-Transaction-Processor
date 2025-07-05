package processor

import (
	"encoding/json"
	"fmt"
	"os"
)

// Snapshot manages account balances and file I/O
type Snapshot struct {
	Balances map[string]int64
}

// LoadSnapshot reads the initial snapshot from JSON
func LoadSnapshot(filePath string) (*Snapshot, error) {
	file, err := os.ReadFile(filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to read snapshot: %w", err)
	}

	var balances map[string]int64
	if err := json.Unmarshal(file, &balances); err != nil {
		return nil, fmt.Errorf("failed to parse snapshot: %w", err)
	}

	// Ensure the validator account exists
	if _, exists := balances["validator"]; !exists {
		balances["validator"] = 0
	}

	return &Snapshot{Balances: balances}, nil
}

// SaveSnapshot saves the current balances to a timestamped file
func (s *Snapshot) SaveSnapshot(timestamp int64, batchIndex int64) error {
	filename := fmt.Sprintf("accounts-%d-%d.json", timestamp, batchIndex)

	data, err := json.MarshalIndent(s.Balances, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal snapshot: %w", err)
	}

	if err := os.WriteFile(filename, data, 0644); err != nil {
		return fmt.Errorf("failed to write snapshot file %s: %w", filename, err)
	}

	return nil
}
