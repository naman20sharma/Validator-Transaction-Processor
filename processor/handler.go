package processor

import (
	"encoding/json"
	"log"
	"net"
	"sync"
	"time"

	"sig-takehome-exercise/models"
)

type TransactionHandler struct {
	processor    *TransactionProcessor
	pendingTxs   []models.Transaction
	pendingMutex sync.Mutex
	stats        TransactionStats
	statsMutex   sync.RWMutex
}

type TransactionStats struct {
	TotalReceived   int64
	TotalProcessed  int64
	TotalInvalid    int64
	TotalFeesPaid   int64
	LastProcessedAt time.Time
}

func NewTransactionHandler(processor *TransactionProcessor) *TransactionHandler {
	return &TransactionHandler{
		processor:  processor,
		pendingTxs: make([]models.Transaction, 0),
		stats: TransactionStats{
			LastProcessedAt: time.Now(),
		},
	}
}

// StartUDPListener starts listening for transactions on UDP port 2001
func (h *TransactionHandler) StartUDPListener() error {
	addr := net.UDPAddr{
		Port: 2001,
		IP:   net.ParseIP("0.0.0.0"),
	}

	conn, err := net.ListenUDP("udp", &addr)
	if err != nil {
		log.Printf("\033[31mFailed to start UDP listener: %v\033[0m", err)
		return err
	}
	defer conn.Close()

	log.Println("\033[34mUDP listener started on port 2001\033[0m")

	// Buffer size matches the specification (max 1000 bytes)
	buf := make([]byte, 1000)

	for {
		n, clientAddr, err := conn.ReadFromUDP(buf)
		if err != nil {
			log.Printf("\033[31mError reading UDP packet: %v\033[0m", err)
			continue
		}

		// Parse transaction
		var tx models.Transaction
		if err := json.Unmarshal(buf[:n], &tx); err != nil {
			h.updateStats(func(s *TransactionStats) {
				s.TotalReceived++
				s.TotalInvalid++
			})
			log.Printf("\033[33mInvalid transaction JSON from %s: %v (raw: %s)\033[0m", clientAddr, err, string(buf[:n]))
			continue
		}

		// Basic validation
		if err := tx.ValidateBasicStructure(); err != nil {
			h.updateStats(func(s *TransactionStats) {
				s.TotalReceived++
				s.TotalInvalid++
			})
			log.Printf("\033[33mInvalid structure from %s: %v (raw: %s)\033[0m", clientAddr, err, string(buf[:n]))
			continue
		}

		h.updateStats(func(s *TransactionStats) {
			s.TotalReceived++
		})

		// Handle transaction validation and queuing in a separate goroutine
		go h.ProcessTransaction(tx, clientAddr)
	}
}

// ProcessTransaction handles a single transaction
func (h *TransactionHandler) ProcessTransaction(tx models.Transaction, clientAddr *net.UDPAddr) {
	// Check if the transaction's fee payer has enough balance,
	// and if instructions are feasible given the current snapshot.
	if !h.processor.ValidateTransaction(tx) {
		h.updateStats(func(s *TransactionStats) {
			s.TotalInvalid++
		})
		txJSON, _ := json.Marshal(tx)
		log.Printf("\033[31mRejected tx from %s: insufficient funds or invalid logic (payer: %s, tx: %s)\033[0m",
			clientAddr, tx.Fee.Payer, string(txJSON))
		return
	}

	// Add to pending transactions for batch processing
	h.pendingMutex.Lock()
	h.pendingTxs = append(h.pendingTxs, tx)
	h.pendingMutex.Unlock()

	log.Printf("\033[32mQueued transaction from %s: %s paying fee %d\033[0m", clientAddr, tx.Fee.Payer, tx.Fee.Amount)
}

// GetPendingTransactions retrieves and clears pending transactions for batch processing
func (h *TransactionHandler) GetPendingTransactions(maxCount int) []models.Transaction {
	h.pendingMutex.Lock()
	defer h.pendingMutex.Unlock()

	if len(h.pendingTxs) == 0 {
		return nil
	}

	usedAccounts := make(map[string]bool)
	var batch []models.Transaction
	var remaining []models.Transaction

	for _, tx := range h.pendingTxs {
		if len(batch) >= maxCount {
			remaining = append(remaining, tx)
			continue
		}

		// Gather all accounts this transaction will touch
		accounts := map[string]struct{}{tx.Fee.Payer: {}}
		for _, instr := range tx.Instructions {
			accounts[instr.Account] = struct{}{}
		}

		// Check for conflicts
		conflict := false
		for acct := range accounts {
			if usedAccounts[acct] {
				conflict = true
				break
			}
		}
		if conflict {
			log.Printf("\033[33mDeferred tx from %s due to account overlap\033[0m", tx.Fee.Payer)
			remaining = append(remaining, tx)
			continue
		}

		// No conflict → add transaction to batch
		for acct := range accounts {
			usedAccounts[acct] = true
		}
		batch = append(batch, tx)
	}

	// Keep remaining txs for future ticks
	h.pendingTxs = remaining
	return batch
}

// UpdateProcessedStats updates statistics after processing a batch
func (h *TransactionHandler) UpdateProcessedStats(count int, totalFees int64) {
	h.updateStats(func(s *TransactionStats) {
		s.TotalProcessed += int64(count)
		s.TotalFeesPaid += totalFees
		s.LastProcessedAt = time.Now()
	})
}

// GetStats returns a copy of current statistics
func (h *TransactionHandler) GetStats() TransactionStats {
	h.statsMutex.RLock()
	defer h.statsMutex.RUnlock()
	return h.stats
}

// updateStats safely updates statistics
func (h *TransactionHandler) updateStats(updater func(*TransactionStats)) {
	h.statsMutex.Lock()
	defer h.statsMutex.Unlock()
	updater(&h.stats)
}

// LogStats periodically logs transaction statistics
func (h *TransactionHandler) LogStats() {
	ticker := time.NewTicker(30 * time.Second)
	defer ticker.Stop()

	for range ticker.C {
		stats := h.GetStats()

		// Calculate duration since last update
		duration := time.Since(stats.LastProcessedAt).Seconds()
		rate := float64(0)
		if duration > 0 {
			rate = float64(stats.TotalFeesPaid) / duration
		}

		log.Printf("\033[36m[Stats] Received: %d, Processed: %d, Invalid: %d, Fees: %d, Rate: %.2f/sec\033[0m",
			stats.TotalReceived, stats.TotalProcessed, stats.TotalInvalid, stats.TotalFeesPaid, rate)
	}
}
