package scheduler

import (
	"reflect"
	"testing"
	"unsafe"
)

func testTaskEntry() {}

func MockInit() {
	taskCount = 0
	currentTask = nil
	nextID = 1
	// Reset tasks array if needed, though taskCount handles the logical reset
	for i := 0; i < MaxTasks; i++ {
		tasks[i] = nil
	}
}

func TestInit(t *testing.T) {
	MockInit()

	Init()

	if taskCount != 1 {
		t.Errorf("Expected taskCount to be 1, got %d", taskCount)
	}

	if tasks[0] == nil {
		t.Fatalf("Expected tasks[0] to be initialized")
	}

	if tasks[0].ID != 0 {
		t.Errorf("Expected tasks[0].ID to be 0, got %d", tasks[0].ID)
	}

	if tasks[0].State != TaskRunning {
		t.Errorf("Expected tasks[0].State to be TaskRunning, got %v", tasks[0].State)
	}

	if currentTask != tasks[0] {
		t.Errorf("Expected currentTask to be tasks[0]")
	}
}

func TestNewTaskEntryBuildsExpectedStack(t *testing.T) {
	MockInit()
	Init()

	const entry uintptr = 0x1122334455667788
	task := NewTaskEntry(entry)
	if task == nil {
		t.Fatalf("Expected task to be created")
	}

	sp := uintptr(task.ESP)
	if sp&15 != 0 {
		t.Fatalf("Expected initial RSP to be 16-byte aligned, got 0x%x", sp)
	}

	gotEntry := *(*uintptr)(unsafe.Pointer(sp + 32))
	if gotEntry != entry {
		t.Fatalf("Expected entry 0x%x, got 0x%x", entry, gotEntry)
	}

	gotFallback := *(*uintptr)(unsafe.Pointer(sp + 40))
	if gotFallback != funcPC(taskAutoExit) {
		t.Fatalf("Expected fallback to taskAutoExit")
	}
}

func TestNewTaskEntryRejectsZeroEntry(t *testing.T) {
	MockInit()
	Init()

	if task := NewTaskEntry(0); task != nil {
		t.Fatalf("Expected nil task for zero entry")
	}
}

func TestFuncPCMatchesReflectPointer(t *testing.T) {
	got := funcPC(testTaskEntry)
	want := reflect.ValueOf(testTaskEntry).Pointer()

	if got == 0 || want == 0 {
		t.Fatalf("Expected non-zero function pointers, got=0x%x want=0x%x", got, want)
	}

	if got != want {
		t.Fatalf("funcPC mismatch: got=0x%x want=0x%x", got, want)
	}
}

func TestNewTaskUsesFunctionEntryPointer(t *testing.T) {
	MockInit()
	Init()

	task := NewTask(testTaskEntry)
	if task == nil {
		t.Fatalf("Expected task to be created")
	}

	sp := uintptr(task.ESP)
	gotEntry := *(*uintptr)(unsafe.Pointer(sp + 32))
	wantEntry := reflect.ValueOf(testTaskEntry).Pointer()

	if gotEntry != wantEntry {
		t.Fatalf("Expected task entry 0x%x, got 0x%x", wantEntry, gotEntry)
	}
}
