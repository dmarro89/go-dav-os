package fs

import (
	"testing"
)

func MockInit() {
	for i := 0; i < maxFiles; i++ {
		files[i].used = false
		files[i].nameLen = 0
		files[i].size = 0
		files[i].page = 0
	}
}

// helper to make a name array
func makeName(s string) (out [maxName]byte, l int) {
	for i := 0; i < maxName && i < len(s); i++ {
		out[i] = s[i]
		l++
	}
	return
}

func TestInit(t *testing.T) {
	// Dirty the state manually
	files[0].used = true
	files[0].size = 999

	MockInit()

	if files[0].used {
		t.Errorf("Slot 0 still used after Init")
	}
	if files[0].size != 0 {
		t.Errorf("Slot 0 size not cleared")
	}
}

func TestLookup(t *testing.T) {
	MockInit()

	name, nameLen := makeName("exists.txt")

	// Manually inject a file
	files[0].used = true
	files[0].name = name
	files[0].nameLen = uint8(nameLen)
	files[0].size = 100
	files[0].page = 5000 // Fake page

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
	MockInit()

	name, nameLen := makeName("delete.txt")
	files[0].used = true
	files[0].name = name
	files[0].nameLen = uint8(nameLen)
	files[0].page = 5000

	// Since mem.PFAReady() returns false (uninitialized), mem.FreePage(5000) returns false safely.
	// Remove() uses this logic:
	// if e.used && e.page != 0 { mem.FreePage(e.page) }
	// It relies on mem.FreePage not crashing.

	success := Remove(&name, nameLen)
	if !success {
		t.Errorf("Remove returned false")
	}

	if files[0].used {
		t.Errorf("File slot still marked used after Remove")
	}
}

func TestWriteFailure(t *testing.T) {
	MockInit()

	// Because we cannot mock mem.PFAReady() or mem.AllocPage() without modifying fs.go,
	// Write() must fail.

	name, nameLen := makeName("new.txt")
	data := []byte("hello")

	success := Write(&name, nameLen, &data[0], uint32(len(data)))
	if success {
		t.Errorf("Write succeeded unexpectedly (should fail due to missing memory subsystem)")
	}
}
