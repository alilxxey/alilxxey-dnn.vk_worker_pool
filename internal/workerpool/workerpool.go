package workerpool

import (
	"log"
	"sync"
)

type Job func() error // TODO: ptr

type Task struct {
	job   Job
	errCh chan error
}

type WorkerPool struct {
	tasks      chan Task
	maxWorkers int
	mu         sync.Mutex
	curWorkers int
	nextID     int
}

func NewWorkerPool(initialWorkers, maxWorkers, chanBuffer int) *WorkerPool {
	pool := &WorkerPool{
		tasks:      make(chan Task, chanBuffer),
		maxWorkers: maxWorkers,
		curWorkers: 0,
		nextID:     1,
	}
	for i := 0; i < initialWorkers; i++ {
		pool.spawnWorker()
	}
	return pool
}

func (p *WorkerPool) SubmitTask(job Job) error {
	task := Task{
		job:   job,
		errCh: make(chan error, 1),
	}
	var spawnID int
	var newCurWorkers int
	p.mu.Lock()
	if p.curWorkers < p.maxWorkers && len(p.tasks) > 1 {
		spawnID = p.nextID
		p.nextID++
		p.curWorkers++
		newCurWorkers = p.curWorkers
	}
	p.mu.Unlock()

	if spawnID != 0 {
		log.Printf("Scaling up: starting worker %d (total workers now %d)\n", spawnID, newCurWorkers)
		go p.worker(spawnID)
	}

	p.tasks <- task
	err := <-task.errCh
	return err
}

func (p *WorkerPool) spawnWorker() {
	p.mu.Lock()
	id := p.nextID
	p.nextID++
	p.curWorkers++
	p.mu.Unlock()

	log.Printf("Starting new worker %d\n", id)
	go p.worker(id)
}

func (p *WorkerPool) worker(id int) {
	for task := range p.tasks {
		log.Printf("Worker %d: processing task", id)
		err := task.job() // TODO: check nil
		log.Printf("Worker %d: completed task; Error: %T", id, err)
		task.errCh <- err
	}
}
