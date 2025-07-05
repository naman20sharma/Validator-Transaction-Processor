package processor

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"time"

	"sig-takehome-exercise/models"
)

const (
	HTTPBatchURL     = "http://localhost:2002/"
	MaxTxPerBatch    = 100
	MaxBatchesPerSec = 100
)

// BatchSender pulls transactions, sends HTTP batches, applies them, and snapshots state.
type BatchSender struct {
	handler      *TransactionHandler
	processor    *TransactionProcessor
	ticker       *time.Ticker
	batchCounter int64
}

// NewBatchSender creates a BatchSender that ticks every second.
func NewBatchSender(handler *TransactionHandler, processor *TransactionProcessor) *BatchSender {
	return &BatchSender{
		handler:   handler,
		processor: processor,
		ticker:    time.NewTicker(time.Second),
	}
}

// Start begins the loop: each second, send up to MaxBatchesPerSec batches.
func (bs *BatchSender) Start() {
	for range bs.ticker.C {
		batchesThisSec := 0
		for batchesThisSec < MaxBatchesPerSec {
			txs := bs.handler.GetPendingTransactions(MaxTxPerBatch)
			if len(txs) == 0 {
				break
			}
			if err := bs.sendAndApplyBatch(txs); err != nil {
				log.Printf("\033[31mBatch send/apply error: %v\033[0m", err)
			}
			batchesThisSec++
		}
	}
}

// sendAndApplyBatch posts the batch to HTTP, then applies each transaction locally,
// deducting fees always and applying instructions only on valid ones.
// It then snapshots the state with the current batch index.
func (bs *BatchSender) sendAndApplyBatch(txs []models.Transaction) error {
	// 1) Send HTTP batch
	body, err := json.Marshal(txs)
	if err != nil {
		return fmt.Errorf("marshal batch: %w", err)
	}
	resp, err := http.Post(HTTPBatchURL, "application/json", bytes.NewBuffer(body))
	if err != nil {
		return fmt.Errorf("post batch: %w", err)
	}
	resp.Body.Close()

	for _, tx := range txs {
		if err := bs.processor.ApplyTransaction(tx); err != nil {
			log.Printf("\033[31mApplyTransaction error (fee applied only): %v\033[0m", err)
		}
	}

	// 3) Increment global batch counter
	bs.batchCounter++
	log.Printf("\033[32m✔ Processed batch #%d → %d transactions, HTTP status: %s\033[0m",
		bs.batchCounter, len(txs), resp.Status)

	// 3.1) Log fee accumulation stats
	totalFees := int64(0)
	for _, tx := range txs {
		totalFees += tx.Fee.Amount
	}

	avgFee := float64(0)
	if len(txs) > 0 {
		avgFee = float64(totalFees) / float64(len(txs))
	}

	log.Printf("\033[36m→ Validator earned %d fees in batch #%d (%.2f avg/tx, %.2f/sec)\033[0m",
		totalFees, bs.batchCounter, avgFee, float64(totalFees))
	// 4) Save snapshot of current balances
	ts := time.Now().Unix()
	if err := bs.processor.GetSnapshot().SaveSnapshot(ts, bs.batchCounter); err != nil {
		log.Printf("\033[31mFailed to save snapshot: %v\033[0m", err)
	}

	return nil
}
