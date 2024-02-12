package workerpool

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"sync"
)

const JSONPlaceholder = "https://jsonplaceholder.typicode.com/posts"

type RealWorker struct {
	NumJobs    int
	NumWorkers int
}

func NewRealWorker(numJobs int, numWorkers int) *RealWorker {
	return &RealWorker{
		NumJobs:    numJobs,
		NumWorkers: numWorkers,
	}
}

type RealWorkerPool struct {
	ResultChan chan User
	Jobs       chan int
	NumWorkers int
	Wg         *sync.WaitGroup
}

func NewRealWorkerPool(numWorkers int, numJobs int) *RealWorkerPool {
	return &RealWorkerPool{
		ResultChan: make(chan User, numJobs),
		Jobs:       make(chan int, numJobs),
		NumWorkers: numWorkers,
		Wg:         &sync.WaitGroup{},
	}
}

type User struct {
	UserID int    `json:"userId"`
	ID     int    `json:"id"`
	Title  string `json:"title"`
	Body   string `json:"body"`
}

func fethUsers(id int) (*User, error) {
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

func (rwp *RealWorkerPool) SubmitJobs(numJobs int) {
	for i := 1; i <= numJobs; i++ {
		rwp.Jobs <- i
	}
	close(rwp.Jobs)
}

func (rwp *RealWorkerPool) StartWorker() {
	for i := 1; i <= rwp.NumWorkers; i++ {
		rwp.Wg.Add(1)
		go func(id int) {
			rwp.Start(id)
		}(i)
	}
}

func (rwp *RealWorkerPool) Start(id int) {
	defer rwp.Wg.Done()
	for job := range rwp.Jobs {
		fmt.Println("worker", id, "started job", job)
		user, err := fethUsers(id)
		if err != nil {
			fmt.Println("failed to fetch user by job", job, "by worker", id)
		}
		if user != nil {
			rwp.ResultChan <- *user
		}
	}
}

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
