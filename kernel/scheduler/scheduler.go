package scheduler

import "unsafe"

const StackSize = 4096
const MaxTasks = 16

type TaskState int

const (
	TaskRunnable TaskState = iota
	TaskRunning
	TaskWaiting
	TaskDead
)

type Task struct {
	ID    int
	ESP   uint64
	State TaskState
	Stack [StackSize]byte
}

var (
	tasks       [MaxTasks]*Task
	taskCount   int
	currentTask *Task
	nextID      int = 1

	// Static allocation for tasks to avoid 'newobject' heap allocation
	taskPool [MaxTasks]Task
)

func Init() {
	taskCount = 0
	nextID = 1
	// Init initial task (0)
	t := &taskPool[0]
	t.ID = 0
	t.State = TaskRunning

	tasks[0] = t
	taskCount = 1
	currentTask = t
}

func NewTask(entry func()) *Task {
	return NewTaskEntry(funcPC(entry))
}

func NewTaskEntry(entry uintptr) *Task {
	if taskCount >= MaxTasks {
		return nil
	}
	if entry == 0 {
		return nil
	}

	idx := taskCount
	t := &taskPool[idx]
	t.ID = nextID
	nextID++
	t.State = TaskRunnable

	// Stack grows down and must match CpuSwitch pop order.
	sp := uintptr(unsafe.Pointer(&t.Stack[0])) + StackSize
	sp &= ^uintptr(15)

	// If entry returns, force task termination instead of jumping to garbage.
	sp -= 8
	*(*uintptr)(unsafe.Pointer(sp)) = funcPC(taskAutoExit)

	// First return target used by CpuSwitch (ret -> entry).
	sp -= 8
	*(*uintptr)(unsafe.Pointer(sp)) = entry

	// CpuSwitch restores RDI, RSI, RBX, RBP from these slots.
	sp -= 32
	*(*uint64)(unsafe.Pointer(sp + 0)) = 0
	*(*uint64)(unsafe.Pointer(sp + 8)) = 0
	*(*uint64)(unsafe.Pointer(sp + 16)) = 0
	*(*uint64)(unsafe.Pointer(sp + 24)) = 0

	t.ESP = uint64(sp)

	tasks[idx] = t
	taskCount++
	return t
}

func taskAutoExit() {
	Exit()
	for {
	}
}

func funcPC(fn func()) uintptr {
	if fn == nil {
		return 0
	}
	fnVal := *(*uintptr)(unsafe.Pointer(&fn))
	if fnVal == 0 {
		return 0
	}
	return *(*uintptr)(unsafe.Pointer(fnVal))
}

func Exit() {
	if currentTask == nil {
		return
	}
	currentTask.State = TaskDead
	Schedule()
	// Should not return if Schedule switched
	for {
	}
}

func Schedule() {
	if taskCount <= 1 {
		return
	}

	oldTask := currentTask

	nextIndex := -1
	currentIndex := -1

	for i := 0; i < taskCount; i++ {
		if tasks[i] == currentTask {
			currentIndex = i
			break
		}
	}

	// Round-robin
	for i := 1; i < taskCount; i++ {
		idx := (currentIndex + i) % taskCount
		if tasks[idx].State == TaskRunnable {
			nextIndex = idx
			break
		}
	}

	if nextIndex == -1 {
		// No runnable task found.
		// If current task is dead, we must find SOMETHING (maybe idle task 0?)
		if oldTask.State == TaskDead {
			// Fallback to task 0 usually, assuming it's permanent
			nextIndex = 0
		} else {
			// Current task is still runnable, just return without switching
			return
		}
	}

	newTask := tasks[nextIndex]

	if oldTask.State != TaskDead {
		oldTask.State = TaskRunnable
	}
	newTask.State = TaskRunning
	currentTask = newTask

	cpuSwitch(&oldTask.ESP, newTask.ESP)
}

func CurrentTaskID() int {
	if currentTask == nil {
		return -1
	}
	return currentTask.ID
}
