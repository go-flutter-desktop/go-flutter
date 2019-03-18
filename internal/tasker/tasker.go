package tasker

// Tasker is a small helper utility that makes it easy to move tasks to a
// different goroutine. This is useful when some work must be executed from a
// specific goroutine / OS thread.
type Tasker struct {
	taskCh chan task
}

type task struct {
	f      func()
	doneCh chan struct{}
}

// New prepares a new Tasker.
func New() *Tasker {
	t := &Tasker{
		taskCh: make(chan task),
	}
	return t
}

// Do runs the given function in the goroutine where ExecuteTasks is called. Do
// blocks until the given function has completed.
func (t *Tasker) Do(f func()) {
	doneCh := make(chan struct{})
	t.taskCh <- task{
		f:      f,
		doneCh: doneCh,
	}
	<-doneCh
}

// ExecuteTasks executes any pending tasks, then returns.
func (t *Tasker) ExecuteTasks() {
	for {
		select {
		case task := <-t.taskCh:
			task.f()
			task.doneCh <- struct{}{}
		default:
			return
		}
	}
}
