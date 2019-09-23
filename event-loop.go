package flutter

import (
	"container/heap"
	"fmt"
	"math"
	"time"

	"github.com/go-flutter-desktop/go-flutter/embedder"
	"github.com/go-flutter-desktop/go-flutter/internal/priorityqueue"
)

// EventLoop is a event loop for the main thread that allows for delayed task
// execution.()
type EventLoop struct {
	// store the task (event) by their priorities
	priorityqueue *priorityqueue.PriorityQueue
	// called when a task has been received, used to Wakeup the rendering event loop
	postEmptyEvent func()

	onExpiredTask func(*embedder.FlutterTask) embedder.Result

	// timeout for non-Rendering events that needs to be processed in a polling manner
	platformMessageRefreshRate time.Duration

	// indetifer for the current thread
	mainThreadID uint64
}

// newEventLoop must ALWAYS be called if the calling goroutine is
// `runtime.LockOSThread()`
func newEventLoop(postEmptyEvent func(), onExpiredTask func(*embedder.FlutterTask) embedder.Result) *EventLoop {
	pq := priorityqueue.NewPriorityQueue()
	heap.Init(pq)
	return &EventLoop{
		priorityqueue:  pq,
		postEmptyEvent: postEmptyEvent,
		onExpiredTask:  onExpiredTask,

		// 25 Millisecond is arbitrary value, not too high (adds too much delay to
		// platform messages) and not too low (heavy CPU consumption).
		// This value isn't related to FPS, as rendering events are process in a
		// waiting manner.
		// Platform message are fetched from the engine every time the rendering
		// event loop process rendering event (e.g.: moving the cursor on the
		// window), when no rendering event occur (e.g., window minimized) platform
		// message are fetch every 25ms.
		platformMessageRefreshRate: time.Duration(25) * time.Millisecond,
	}
}

// RunOnCurrentThread FlutterDocs:
//   May be called from any thread. Should return true if tasks posted on the
//   calling thread will be run on that same thread.
//
// The functions PostTask and onExpiredTask should be called from the same
// thread, this is ensured if the creation of the event loop (through
// `newEventLoop`) and the PostTask callback (through
// `a.engine.TaskRunnerPostTask = eventLoop.PostTask`) are done on a calling
// goroutine which always execute in that thread (`runtime.LockOSThread()`).
func (t *EventLoop) RunOnCurrentThread() bool {
	return true
}

// PostTask posts a Flutter engine tasks to the event loop for delayed execution.
// PostTask must ALWAYS be called on the same goroutine/thread as `newEventLoop`
func (t *EventLoop) PostTask(task embedder.FlutterTask, targetTimeNanos uint64) {

	taskDuration := time.Duration(targetTimeNanos) * time.Nanosecond
	engineDuration := time.Duration(embedder.FlutterEngineGetCurrentTime())

	t.priorityqueue.Lock()
	item := &priorityqueue.Item{
		Value:    task,
		FireTime: time.Now().Add(taskDuration - engineDuration),
	}
	heap.Push(t.priorityqueue, item)
	t.priorityqueue.Unlock()

	t.postEmptyEvent()
}

// WaitForEvents waits for an any Rendering or pending Flutter Engine events
// and returns when either is encountered.
// Expired engine events are processed
func (t *EventLoop) WaitForEvents(rendererWaitEvents func(float64)) {
	now := time.Now()

	expiredTasks := make([]*priorityqueue.Item, 0)
	var top *priorityqueue.Item

	t.priorityqueue.Lock()
	for t.priorityqueue.Len() > 0 {

		// Remove the item from the delayed tasks queue.
		top = heap.Pop(t.priorityqueue).(*priorityqueue.Item)

		// If this task (and all tasks after this) has not yet expired, there is
		// nothing more to do. Quit iterating.
		if top.FireTime.After(now) {
			heap.Push(t.priorityqueue, top) // push the item back into the queue
			break
		}

		// Make a record of the expired task. Do NOT service the task here
		// because we are still holding onto the task queue mutex. We don't want
		// other threads to block on posting tasks onto this thread till we are
		// done processing expired tasks.
		expiredTasks = append(expiredTasks, top)

	}
	hasTask := t.priorityqueue.Len() != 0
	t.priorityqueue.Unlock()

	// Fire expired tasks.
	for _, item := range expiredTasks {
		task := item.Value
		if t.onExpiredTask(&task) != embedder.ResultSuccess {
			fmt.Printf("go-flutter: couldn't process task %v\n", task)
		}
	}

	// Sleep till the next task needs to be processed. If a new task comes
	// along, the rendererWaitEvents will be resolved early because PostTask
	// posts an empty event.
	if !hasTask {
		rendererWaitEvents(t.platformMessageRefreshRate.Seconds())
	} else {
		if top.FireTime.After(now) {
			durationWait := math.Min(top.FireTime.Sub(now).Seconds(), t.platformMessageRefreshRate.Seconds())
			rendererWaitEvents(durationWait)
		}
	}
}
