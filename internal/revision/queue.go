package revision

import (
	"fmt"
	"sync"
	"time"

	"github.com/paradoxe35/myreviser-go/internal/logger"
)

// RevisionTask represents a single revision task
type RevisionTask struct {
	ID        string
	Text      string
	Mode      string // "select_all" or "selection"
	Timestamp time.Time
	Result    chan RevisionResult
}

// RevisionResult represents the result of a revision
type RevisionResult struct {
	RevisedText string
	Error       error
}

// Queue manages revision tasks
type Queue struct {
	mu       sync.Mutex
	tasks    chan *RevisionTask
	current  *RevisionTask
	stopChan chan struct{}
	stopped  bool
}

// NewQueue creates a new revision queue
func NewQueue() *Queue {
	return &Queue{
		tasks:    make(chan *RevisionTask, 10), // Buffer up to 10 tasks
		stopChan: make(chan struct{}),
	}
}

// Add adds a new task to the queue
func (q *Queue) Add(task *RevisionTask) error {
	q.mu.Lock()
	defer q.mu.Unlock()

	if q.stopped {
		return fmt.Errorf("queue is stopped")
	}

	select {
	case q.tasks <- task:
		logger.Info("Task added to queue", "id", task.ID, "mode", task.Mode)
		return nil
	default:
		return fmt.Errorf("queue is full")
	}
}

// GetCurrent returns the current task
func (q *Queue) GetCurrent() *RevisionTask {
	q.mu.Lock()
	defer q.mu.Unlock()
	return q.current
}

// SetCurrent sets the current task
func (q *Queue) SetCurrent(task *RevisionTask) {
	q.mu.Lock()
	defer q.mu.Unlock()
	q.current = task
}

// Clear clears the current task
func (q *Queue) Clear() {
	q.mu.Lock()
	defer q.mu.Unlock()
	q.current = nil
}

// IsProcessing returns whether a task is currently being processed
func (q *Queue) IsProcessing() bool {
	q.mu.Lock()
	defer q.mu.Unlock()
	return q.current != nil
}

// Stop stops the queue
func (q *Queue) Stop() {
	q.mu.Lock()
	defer q.mu.Unlock()

	if !q.stopped {
		q.stopped = true
		close(q.stopChan)
		close(q.tasks)
	}
}

// ProcessTasks starts processing tasks from the queue
func (q *Queue) ProcessTasks(processor func(*RevisionTask) RevisionResult) {
	for {
		select {
		case task := <-q.tasks:
			if task == nil {
				continue
			}

			q.SetCurrent(task)
			logger.Info("Processing task", "id", task.ID, "mode", task.Mode)

			// Process the task
			result := processor(task)

			// Send result back
			select {
			case task.Result <- result:
			case <-time.After(1 * time.Second):
				logger.Warn("Task result timeout", "id", task.ID)
			}

			q.Clear()

		case <-q.stopChan:
			logger.Info("Queue processor stopped")
			return
		}
	}
}

// GetQueueSize returns the number of tasks in the queue
func (q *Queue) GetQueueSize() int {
	return len(q.tasks)
}