package workerpool

import (
	"errors"
	"log"
	"sync"
	"sync/atomic"
	"time"
)

type Job[T any] func() (T, error)

// Pool представляет интерфейс пула воркеров
type Pool[T any] interface {
	// SubmitTask отправляет задачу на выполнение воркерам
	SubmitTask(Job[T]) (T, error)
	// Shutdown останавливает прием задач и ждет освобождения всех воркеров
	Shutdown()
}

// WorkerPool управляет пулом воркеров для выполнения Job
type WorkerPool[T any] struct {
	tasks       chan Task[T]
	maxWorkers  int
	minWorkers  int
	workerCount atomic.Uint32
	workerID    atomic.Uint32
	closed      atomic.Bool
	wg          sync.WaitGroup
}

// Task хранит задачу и канал для возвращения ошибки
type Task[T any] struct {
	job      Job[T]
	errCh    chan error
	resultCh chan T
}

// NewWorkerPool создаёт пул с [initialWorkers] воркерами,
// максимальным числом maxWorkers и буфером задач размером chanBuffer
func NewWorkerPool[T any](minWorkers, maxWorkers, chanBuffer int) *WorkerPool[T] {
	pool := &WorkerPool[T]{
		tasks:      make(chan Task[T], chanBuffer),
		maxWorkers: maxWorkers,
		minWorkers: minWorkers,
	}
	for i := 0; i < minWorkers; i++ {
		pool.spawnWorker()
	}
	return pool
}

// SubmitTask отправляет задачу в очередь и ждёт её выполнения
func (p *WorkerPool[T]) SubmitTask(job Job[T]) (T, error) {
	if p.closed.Load() {
		var zero T
		return zero, errors.New("worker pool is down")
	}
	task := Task[T]{
		job:      job,
		errCh:    make(chan error, 1),
		resultCh: make(chan T, 1),
	}

	if p.workerCount.Load() < uint32(p.maxWorkers) && len(p.tasks) > 1 {
		p.spawnWorker()
	}
	p.tasks <- task

	if err := <-task.errCh; err != nil {
		var smth T
		return smth, err
	}

	return <-task.resultCh, nil
}

// Shutdown корректно завершает приём новых задач и ждёт завершения текущих.
func (p *WorkerPool[T]) Shutdown() {
	if p.closed.CompareAndSwap(false, true) {
		close(p.tasks)
	}
	p.wg.Wait()
}

// spawnWorker запускает нового воркера с уникальным idd
func (p *WorkerPool[T]) spawnWorker() {
	// атомарные операции чтобы не использовать мьютекс
	id := p.workerID.Add(1)
	p.workerCount.Add(1)
	p.wg.Add(1)
	log.Printf("Starting new worker with id %d", id)

	go p.worker(id)
}

// worker обрабатывает задачи из канала tasks до его закрытия, автоматически down-скейлится если нет задач
func (p *WorkerPool[T]) worker(id uint32) {
	defer func() {
		if r := recover(); r != nil {
			log.Printf("Worker %d recovered from panic: %v", id, r)
		}
		p.workerCount.Add(^uint32(0))
		p.wg.Done()
		log.Printf("Worker %d: stopped", id)
	}()

	idle := time.NewTimer(30 * time.Second)
	defer idle.Stop()

	for {
		select {
		case task, ok := <-p.tasks:
			if !ok {
				return
			}
			if !idle.Stop() {
				<-idle.C
			}
			idle.Reset(30 * time.Second)

			var (
				result T
				err    error
			)
			func() {
				defer func() {
					if r := recover(); r != nil {
						err = errors.New("job panic")
						log.Printf("Worker %d: job panicked: %v", id, r)
					}
				}()
				result, err = task.job()
			}()

			task.errCh <- err
			task.resultCh <- result
			close(task.errCh)
			close(task.resultCh)

		case <-idle.C:
			if int(p.workerCount.Load()) > p.minWorkers {
				log.Printf("Worker %d: idle timeout; exiting", id)
				return
			}
			idle.Reset(30 * time.Second)
		}
	}
}
