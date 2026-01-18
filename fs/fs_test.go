package fs

import (
	"testing"
)

// helper to make a name array
func makeName(s string) (out [maxName]byte, l int) {
	for i := 0; i < maxName && i < len(s); i++ {
		out[i] = s[i]
		l++
	}
	return
}

func TestInit(t *testing.T) {
	// Dirty the state manually since we are in package fs
	files[0].used = true
	files[0].size = 999
	files[0].page = 123

	Init()

	for i := 0; i < maxFiles; i++ {
		if files[i].used {
			t.Errorf("Slot %d still used after Init", i)
		}
		if files[i].size != 0 {
			t.Errorf("Slot %d size not cleared", i)
		}
	}
}

func TestLookup(t *testing.T) {
	Init()

	name, nameLen := makeName("exists.txt")

	// Manually inject a file
	files[0].used = true
	files[0].name = name
	files[0].nameLen = uint8(nameLen)
	files[0].size = 100
	files[0].page = 5000 // Arbitrary

	// Test positive lookup
	page, size, ok := Lookup(&name, nameLen)
	if !ok {
		t.Errorf("Lookup failed for injected file")
	}
	if page != 5000 {
		t.Errorf("Expected page 5000, got %d", page)
	}
	if size != 100 {
		t.Errorf("Expected size 100, got %d", size)
	}

	// Test negative lookup
	missing, missingLen := makeName("missing.txt")
	_, _, ok = Lookup(&missing, missingLen)
	if ok {
		t.Errorf("Lookup succeeded for missing file")
	}
}

func TestRemove(t *testing.T) {
	Init()

	name, nameLen := makeName("delete.txt")

	// Inject file
	files[0].used = true
	files[0].name = name
	files[0].nameLen = uint8(nameLen)
	files[0].page = 100 // valid page index, but mem.PFAReady() is false, so FreePage returns safe

	// Test Remove
	// Note: fs.go calls mem.FreePage.
	// Since mem.PFAReady() defaults to false in a fresh process without InitPFA call,
	// mem.FreePage(100) will return false immediately and NOT crash.
	success := Remove(&name, nameLen)
	if !success {
		t.Errorf("Remove returned false")
	}

	if files[0].used {
		t.Errorf("File slot still marked used after Remove")
	}
	if files[0].page != 0 {
		t.Errorf("Page not cleared")
	}

	// Test removing non-existent
	success = Remove(&name, nameLen)
	if success {
		t.Errorf("Remove succeeded for already deleted file")
	}
}

func TestWriteFailure(t *testing.T) {
	Init()

	// Since we cannot initialize the memory subsystem (mem) without a real kernel environment,
	// Write() is expected to fail.
	// This test confirms it handles the failure gracefully (returns false).

	name, nameLen := makeName("new.txt")
	data := []byte("hello")

	success := Write(&name, nameLen, &data[0], uint32(len(data)))
	if success {
		t.Errorf("Write succeeded but should have failed due to uninitialized memory")
	}

	// Verify no slot was used (or it was cleaned up/not committed)
	// fs.go Write logic: finds slot, THEN allocates page.
	// if mem.PFAReady() is false, it returns false inside the `if !e.used` block.
	// So e.used might remain false?
	// Actually logic is:
	// idx = findFreeSlot()
	// e := &files[idx]
	// if !e.used {
	//    if !mem.PFAReady() { return false }
	//    ...
	//    e.used = true
	// }
	// So 'used' flag is NOT set if PFA not ready. Correct.

	if files[0].used {
		t.Errorf("File slot marked used despite Write failure")
	}
}

func TestMaxFilesEntries(t *testing.T) {
	Init()

	// Fill table
	for i := 0; i < maxFiles; i++ {
		files[i].used = true
	}

	if findFreeSlot() != -1 {
		t.Errorf("Should return -1 when full")
	}

	// With full table, Write should return false immediately (after finding no name match and no free slot)
	name, nameLen := makeName("overflow.txt")
	data := []byte("x")
	success := Write(&name, nameLen, &data[0], 1)
	if success {
		t.Errorf("Write should fail when full")
	}
}
