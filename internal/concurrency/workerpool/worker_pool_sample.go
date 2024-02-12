package workerpool

import (
	"fmt"
	"time"
)

// Problem Statement:
//
// Implement a worker pool pattern in Go. The worker pool should have a fixed
// number of workers. The main program should submit a number of jobs to the
// worker pool, and the workers should process these jobs concurrently. Each
// job should take a fixed amount of time to process. After all jobs are
// processed, the main program should wait for all results to be collected
// from the workers.

// Workers represents a group of workers with the number of jobs and workers.
type Workers struct {
	NumJobs    int // Number of jobs to be processed.
	NumWorkers int // Number of workers to process the jobs.
}

// NewWorker creates a new Workers instance with the given number of jobs and workers.
func NewWorker(numJobs int, numWorkers int) *Workers {
	return &Workers{NumJobs: numJobs, NumWorkers: numWorkers}
}

// WorkerPool represents a pool of workers.
type WorkerPool struct {
	Jobs       chan int // Channel to send jobs to workers.
	Results    chan int // Channel to receive results from workers.
	NumWorkers int      // Number of workers in the pool.
}

// NewWorkerPool creates a new worker pool with the specified number of workers and buffer size for jobs and results.
func NewWorkerPool(numWorkers int, numJobs int) *WorkerPool {
	return &WorkerPool{
		Jobs:       make(chan int, numJobs),
		Results:    make(chan int, numJobs),
		NumWorkers: numWorkers,
	}
}

// StartWorker starts the worker pool with the specified number of workers.
func (wp *WorkerPool) StartWorker() {
	for i := 1; i <= wp.NumWorkers; i++ {
		go func(id int) {
			wp.runJob(id)
		}(i)
	}
}

// runJob represents the logic for a worker to process jobs.
func (wp *WorkerPool) runJob(id int) {
	for job := range wp.Jobs {
		fmt.Println("Worker", id, "started job", job)
		time.Sleep(2 * time.Second)
		fmt.Println("Worker", id, "finished job", job)
		wp.Results <- job * 2
	}
}

// SubmitJob submits the specified number of jobs to the worker pool.
func (wp *WorkerPool) SubmitJob(numJobs int) {
	for i := 1; i <= numJobs; i++ {
		wp.Jobs <- i
	}
	close(wp.Jobs)
}

// PrintResult prints the results received from the worker pool.
func (wp *WorkerPool) PrintResult(numJobs int) {
	for i := 1; i <= numJobs; i++ {
		val := <-wp.Results
		fmt.Println(val)
	}
}

// Run runs the workers with the configured number of jobs and workers.
func (w *Workers) Run() {
	workerPool := NewWorkerPool(w.NumWorkers, w.NumJobs)
	workerPool.StartWorker()
	workerPool.SubmitJob(w.NumJobs)
	workerPool.PrintResult(w.NumJobs)
}
