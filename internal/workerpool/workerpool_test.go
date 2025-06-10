package workerpool

import (
	"errors"
	"sync"
	"testing"
	"time"
)

func TestSubmitTask(t *testing.T) {
	pool := NewWorkerPool[int](2, 4, 10)
	defer pool.Shutdown()

	got, err := pool.SubmitTask(func() (int, error) { return 40 + 2, nil })
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got != 42 {
		t.Fatalf("want 42, got %d", got)
	}
}

func TestConcurrentScaleUp(t *testing.T) {
	const (
		initial   = 1
		max       = 4
		taskCount = 10
		sleep     = 50 * time.Millisecond
	)

	pool := NewWorkerPool[int](initial, max, taskCount)
	defer pool.Shutdown()

	var wg sync.WaitGroup
	wg.Add(taskCount)

	start := time.Now()
	for i := 0; i < taskCount; i++ {
		go func() {
			defer wg.Done()
			_, _ = pool.SubmitTask(func() (int, error) {
				time.Sleep(sleep)
				return 1, nil
			})
		}()
	}
	wg.Wait()
	elapsed := time.Since(start)

	serial := time.Duration(taskCount) * sleep
	if elapsed >= serial/2 {
		t.Fatalf("pool не масштабируется: elapsed %s, expect < %s", elapsed, serial/2)
	}
}

func TestJobPanicDoesNotKillWorker(t *testing.T) {
	pool := NewWorkerPool[int](2, 4, 4)
	defer pool.Shutdown()

	_, err := pool.SubmitTask(func() (int, error) {
		panic("boom")
	})
	if err == nil || !errors.Is(err, errors.New("job panic")) && err.Error() != "job panic" {
		t.Fatalf("want 'job panic' error, got %v", err)
	}

	got, err := pool.SubmitTask(func() (int, error) { return 7, nil })
	if err != nil || got != 7 {
		t.Fatalf("воркер не восстановился: result=%d err=%v", got, err)
	}
}

func TestSubmitAfterShutdown(t *testing.T) {
	pool := NewWorkerPool[int](1, 2, 2)
	pool.Shutdown()

	_, err := pool.SubmitTask(func() (int, error) { return 0, nil })
	if err == nil {
		t.Fatal("ожидалась ошибка после Shutdown, но её нет")
	}
}
