package main

import (
	"fmt"
	"math/rand"
	"strings"
	"sync"
	"time"
)

type LogEntry struct {
	ID          int
	RawLine     string
	UserID      int
	StatusCode  int
	IsValid     bool
	IsProcessed bool
}

const (
	numLogLines = 200
	numWorkers  = 5
)

func logGenerator(count int) <-chan LogEntry {
	output := make(chan LogEntry, 10)
	rand.New(rand.NewSource(time.Now().UnixNano()))

	go func() {
		defer close(output)
		fmt.Println("--- Stage 1: Log Generator STARTED ---")
		for i := 1; i <= count; i++ {
			statusCode := 200
			if rand.Intn(10) < 2 {
				statusCode = 404
				if rand.Intn(10) == 0 {
					output <- LogEntry{ID: i, RawLine: "MALFORMED_DATA_BROKEN_JSON", StatusCode: 0}
					continue
				}
			}

			entry := LogEntry{
				ID:      i,
				UserID:  rand.Intn(1000) + 1,
				RawLine: fmt.Sprintf(`{"time": "%s", "status": %d, "user": %d, "path": "/api/data/%d"}`, time.Now().Format(time.RFC3339), statusCode, i, i),
			}

			output <- entry
			time.Sleep(time.Microsecond * 100)
		}
		fmt.Println("--- Stage 1: Log Generator FINISHED ---")
	}()

	return output
}

func logValidator(input <-chan LogEntry) <-chan LogEntry {
	output := make(chan LogEntry, 10)

	go func() {
		defer close(output)
		fmt.Println("--- Stage 2: Log Validator STARTED ---")
		for entry := range input {
			if strings.HasPrefix(entry.RawLine, "MALFORMED") {
				entry.IsValid = false
				fmt.Printf("Validator: Filtered malformed log ID %d\n", entry.ID)
			} else {
				entry.IsValid = true
			}

			output <- entry
		}
		fmt.Println("--- Stage 2: Log Validator FINISHED ---")
	}()

	return output
}

func logProcessor(input <-chan LogEntry, numWorkers int) <-chan LogEntry {
	output := make(chan LogEntry, 10)
	var wg sync.WaitGroup

	worker := func(id int) {
		defer wg.Done()
		for entry := range input {
			if !entry.IsValid {
				// Bypass processing for invalid entries, but still pass to the next stage if needed
				output <- entry
				continue
			}

			sleepDuration := time.Duration(rand.Intn(100)+50) * time.Millisecond
			time.Sleep(sleepDuration)

			entry.IsProcessed = true
			fmt.Printf("Processor | Worker %d completed processing for log ID %d (Status %d)\n", id, entry.ID, entry.StatusCode)
			output <- entry
		}
	}

	for i := 1; i <= numWorkers; i++ {
		wg.Add(1)
		go worker(i)
	}

	go func() {
		wg.Wait()
		close(output)
		fmt.Println("--- Stage 3: Log Processor FINISHED ---")
	}()

	return output
}

func logAggregator(input <-chan LogEntry) {
	fmt.Println("\n--- Stage 4: Log Aggregator STARTED ---")
	totalCount := 0
	validCount := 0
	processedCount := 0
	statusCounts := make(map[int]int)

	for entry := range input {
		totalCount++

		if entry.IsValid {
			validCount++
		}
		if entry.IsProcessed {
			processedCount++
		}

		statusCounts[entry.StatusCode]++
	}

	fmt.Println("--- Stage 4: Log Aggregator FINISHED ---")

	fmt.Println("\n==============================================")
	fmt.Println("           LOG ANALYTICS REPORT")
	fmt.Println("==============================================")
	fmt.Printf("Total Logs Generated: %d\n", totalCount)
	fmt.Printf("Logs Filtered by Validator: %d\n", totalCount-validCount)
	fmt.Printf("Logs Passed to Processor: %d\n", validCount)
	fmt.Printf("Logs Successfully Processed: %d\n", processedCount)
	fmt.Println("\n--- Status Code Summary ---")
	for status, count := range statusCounts {
		fmt.Printf("  Status %d: %d entries\n", status, count)
	}
	fmt.Println("==============================================")
}

func main() {
	startTime := time.Now()

	generatorOuput := logGenerator(numLogLines)

	validatorOutput := logValidator(generatorOuput)

	processorOutput := logProcessor(validatorOutput, numWorkers)

	logAggregator(processorOutput)

	fmt.Printf("\nPipeline executed successfully in %v\n", time.Since(startTime))
}
