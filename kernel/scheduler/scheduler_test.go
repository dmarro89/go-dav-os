package scheduler

import "testing"

func MockInit() {
	taskCount = 0
	currentTask = nil
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
