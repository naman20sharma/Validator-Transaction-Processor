package main

import (
	"log"
	"os"

	"sig-takehome-exercise/processor"
)

func main() {
	if len(os.Args) != 2 {
		log.Fatalf("Usage: %s <path/to/accounts.json>", os.Args[0])
	}
	snapshotPath := os.Args[1]

	// 1) Load initial snapshot
	snap, err := processor.LoadSnapshot(snapshotPath)
	if err != nil {
		log.Fatalf("Load snapshot error: %v", err)
	}
	log.Printf("\033[32m Snapshot loaded from %s\033[0m", snapshotPath)
	// 2) Initialize processor, handler, and batch sender
	proc := processor.NewTransactionProcessor(snap)
	handler := processor.NewTransactionHandler(proc)
	batcher := processor.NewBatchSender(handler, proc)

	// 3) Start UDP listener
	go func() {
		if err := handler.StartUDPListener(); err != nil {
			log.Fatalf("UDP listener error: %v", err)
		}
	}()

	// 4) Log stats periodically
	go handler.LogStats()

	log.Println("\033[32m Validator is up and listening for transactions on port 2001...\033[0m")
	// 5) Start batching + snapshot loop
	batcher.Start()

}
