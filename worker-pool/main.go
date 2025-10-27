package main

import (
	"fmt"
	"math/rand"
	"sync"
	"time"
)

type Job struct {
	ID        int
	InputData int
}

type Result struct {
	JobID       int
	WorkerID    int
	OutputValue int
	Duration    time.Duration
	Err         error
}

func worker(id int, jobs <-chan Job, results chan<- Result, wg *sync.WaitGroup) {
	defer wg.Done()
	fmt.Printf("Worker %d started.\n", id)

	for job := range jobs {
		startTime := time.Now()

		sleepDuration := time.Duration(rand.Intn(500)+100) * time.Millisecond
		time.Sleep(sleepDuration)

		var err error
		output := job.InputData * 2
		if rand.Intn(10) == 0 {
			err = fmt.Errorf("simulated error: processing job %d failed due to transient issue", job.ID)
			output = 0
		}

		result := Result{
			JobID:       job.ID,
			WorkerID:    id,
			OutputValue: output,
			Duration:    time.Since(startTime),
			Err:         err,
		}

		if result.Err != nil {
			fmt.Printf("Worker %d FAILED job %d (Input: %d) in %v. Error: %s\n", id, job.ID, job.InputData, result.Duration, result.Err)
		} else {
			fmt.Printf("Worker %d completed job %d (Input: %d, Output: %d) in %v\n", id, job.ID, job.InputData, result.OutputValue, result.Duration)
		}

		results <- result
	}

	fmt.Printf("Worker %d gracefully shut down.\n", id)
}

func main() {
	const numJobs = 50
	const numWorkers = 5
	const channelCapacity = 10

	jobs := make(chan Job, channelCapacity)
	results := make(chan Result, numJobs)

	var workerWG sync.WaitGroup
	rand.New(rand.NewSource(time.Now().UnixNano()))

	fmt.Println("--- Starting Worker Pool ---")
	fmt.Printf("Total Jobs: %d, Max Workers: %d\n\n", numJobs, numWorkers)

	for i := 1; i <= numWorkers; i++ {
		workerWG.Add(1)
		go worker(i, jobs, results, &workerWG)
	}

	go func() {
		for j := 1; j <= numJobs; j++ {
			jobs <- Job{
				ID:        j,
				InputData: rand.Intn(100) + 1,
			}
		}

		close(jobs)
	}()

	var collectorWG sync.WaitGroup
	collectorWG.Add(1)
	allResults := make([]Result, 0, numJobs)

	go func() {
		defer collectorWG.Done()
		fmt.Println("\n--- Result Collector started ---")
		for r := range results {
			allResults = append(allResults, r)
		}
		fmt.Println("--- Result Collector finished ---")
	}()

	workerWG.Wait()

	close(results)

	collectorWG.Wait()

	successfulJobs := 0
	failedJobs := 0
	totalDuration := time.Duration(0)

	for _, res := range allResults {
		totalDuration += res.Duration
		if res.Err == nil {
			successfulJobs++
		} else {
			failedJobs++
		}
	}

	fmt.Println("\n==============================================")
	fmt.Println("           WORKER POOL SUMMARY")
	fmt.Println("==============================================")
	fmt.Printf("Total Jobs Submitted: %d\n", numJobs)
	fmt.Printf("Total Workers Used: %d\n", numWorkers)
	fmt.Printf("Successful Jobs: %d\n", successfulJobs)
	fmt.Printf("Failed Jobs: %d\n", failedJobs)
	fmt.Printf("Avg Job Duration: %v\n", totalDuration/time.Duration(numJobs))
	fmt.Println("==============================================")
	fmt.Println("Program finished gracefully.")

}
