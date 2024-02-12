// Package workerpool provides a solution for fetching user data from JSONPlaceholder using a worker pool.
//
// Problem Statement:
//
// We have a requirement to fetch user data from a REST API endpoint (JSONPlaceholder - https://jsonplaceholder.typicode.com/posts)
// where each user is identified by an ID. We want to fetch data for multiple users concurrently to minimize the overall time
// required to fetch all the data. To achieve this, we implement a worker pool pattern where multiple workers fetch user data
// concurrently. Each worker fetches user data for a specific user ID and sends the fetched data to a result channel. The main
// goroutine reads the data from the result channel and writes it to a JSON file. The number of workers and the number of user
// IDs to fetch data for are configurable.
package workerpool

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"sync"
)

// JSONPlaceholder represents the base URL for fetching users' data.
const JSONPlaceholder = "https://jsonplaceholder.typicode.com/posts"

// RealWorker represents a group of workers with the number of jobs and workers.
type RealWorker struct {
	NumJobs    int // Number of jobs to be processed.
	NumWorkers int // Number of workers to process the jobs.
}

// NewRealWorker creates a new RealWorker instance with the given number of jobs and workers.
func NewRealWorker(numJobs int, numWorkers int) *RealWorker {
	return &RealWorker{
		NumJobs:    numJobs,
		NumWorkers: numWorkers,
	}
}

// RealWorkerPool represents a pool of workers fetching user data from JSONPlaceholder.
type RealWorkerPool struct {
	ResultChan chan User       // Channel to receive fetched user data.
	Jobs       chan int        // Channel to send jobs to workers.
	NumWorkers int             // Number of workers in the pool.
	Wg         *sync.WaitGroup // WaitGroup to synchronize worker goroutines.
}

// NewRealWorkerPool creates a new worker pool with the specified number of workers and jobs.
func NewRealWorkerPool(numWorkers int, numJobs int) *RealWorkerPool {
	return &RealWorkerPool{
		ResultChan: make(chan User, numJobs),
		Jobs:       make(chan int, numJobs),
		NumWorkers: numWorkers,
		Wg:         &sync.WaitGroup{},
	}
}

// User represents a user fetched from JSONPlaceholder.
type User struct {
	UserID int    `json:"userId"`
	ID     int    `json:"id"`
	Title  string `json:"title"`
	Body   string `json:"body"`
}

// fetchUsers fetches a user from JSONPlaceholder by its ID.
func fetchUsers(id int) (*User, error) {
	response, err := http.Get(fmt.Sprintf("%s/%d", JSONPlaceholder, id))
	if err != nil {
		return nil, err
	}
	defer response.Body.Close()

	var user User
	if err = json.NewDecoder(response.Body).Decode(&user); err != nil {
		return nil, err
	}
	return &user, nil
}

// SubmitJobs submits the specified number of jobs to the worker pool.
func (rwp *RealWorkerPool) SubmitJobs(numJobs int) {
	for i := 1; i <= numJobs; i++ {
		rwp.Jobs <- i
	}
	close(rwp.Jobs)
}

// StartWorker starts the worker pool with the specified number of workers.
func (rwp *RealWorkerPool) StartWorker() {
	for i := 1; i <= rwp.NumWorkers; i++ {
		rwp.Wg.Add(1)
		go func(id int) {
			rwp.Start(id)
		}(i)
	}
}

// Start starts a worker to fetch and process user data.
func (rwp *RealWorkerPool) Start(id int) {
	defer rwp.Wg.Done()
	for job := range rwp.Jobs {
		fmt.Println("worker", id, "started job", job)
		user, err := fetchUsers(id)
		if err != nil {
			fmt.Println("failed to fetch user by job", job, "by worker", id)
		}
		if user != nil {
			rwp.ResultChan <- *user
		}
	}
}

// WriteResultToFile writes the fetched user data to a JSON file.
func (rwp *RealWorkerPool) WriteResultToFile(numJobs int) {
	file, err := os.Create("results.json")
	if err != nil {
		fmt.Println(err)
	}
	defer file.Close()
	encoder := json.NewEncoder(file)
	for i := 1; i <= numJobs; i++ {
		user := <-rwp.ResultChan
		err = encoder.Encode(user)
		if err != nil {
			fmt.Println(err)
		}
		fmt.Println("fetched user :", user.ID)
	}
}

// Run runs the RealWorker with the configured number of jobs and workers.
func (rw *RealWorker) Run() {
	pool := NewRealWorkerPool(rw.NumWorkers, rw.NumJobs)
	pool.SubmitJobs(rw.NumJobs)

	pool.StartWorker()

	go func() {
		pool.Wg.Wait()
		close(pool.ResultChan)
	}()
	pool.WriteResultToFile(rw.NumJobs)
}
