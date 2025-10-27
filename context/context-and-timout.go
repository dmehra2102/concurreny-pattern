package main

import (
	"context"
	"fmt"
	"sync"
	"time"
)

func performTask(ctx context.Context, id int, duration time.Duration, wg *sync.WaitGroup) {
	defer wg.Done()

	fmt.Printf("[Task %d] Started processing (simulated duration: %v)\n", id, duration)

	select {
	case <-time.After(duration):
		fmt.Printf("[Task %d] ✅ Completed successfully after %v.\n", id, duration)
	case <-ctx.Done():
		fmt.Printf("[Task %d] 🛑 Canceled early! Reason: %v\n", id, ctx.Err())
		return
	}
}

func main() {
	var wg sync.WaitGroup
	rootCtx := context.Background()

	const timeoutDuration = 200 * time.Millisecond
	fmt.Println("--- SCENARIO 1: Strict Timeout ---")

	timeoutCtx, timeoutCancel := context.WithTimeout(rootCtx, timeoutDuration)
	defer timeoutCancel()

	wg.Add(1)
	go performTask(timeoutCtx, 1, 100*time.Millisecond, &wg)

	wg.Add(1)
	go performTask(timeoutCtx, 2, 500*time.Millisecond, &wg)

	fmt.Println("\n--- SCENARIO 2: Manual Cancellation ---")
	// Create a context that needs to be manually canceled.
	manualCtx, manualCancel := context.WithCancel(rootCtx)

	wg.Add(1)
	go performTask(manualCtx, 3, 50*time.Millisecond, &wg)

	wg.Add(1)
	go performTask(manualCtx, 4, 1000*time.Millisecond, &wg)

	fmt.Println("Main: Waiting 100ms, then sending manual cancellation signal...")
	time.Sleep(100 * time.Millisecond)

	manualCancel()
	fmt.Println("Main: Cancellation signal SENT.")

	// Wait for all four tasks to complete (or be canceled).
	wg.Wait()
	fmt.Println("\n--- All goroutines finished/canceled. Program exit. ---")
}
