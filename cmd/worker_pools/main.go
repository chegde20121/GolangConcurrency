package main

import (
	"github.com/chegde20121/GolangConcurrency/internal/concurrency/workerpool"
)

func main() {
	//worker := workerpool.NewWorker(5, 3)
	//worker.Run()

	//Real-Time Example
	worker_group := workerpool.NewRealWorker(5, 5)
	worker_group.Run()
}
