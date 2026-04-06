package kernel

import (
	"github.com/dmarro89/go-dav-os/kernel/scheduler"
	"github.com/dmarro89/go-dav-os/keyboard"
	"github.com/dmarro89/go-dav-os/keyboard/layout"
)

var currentLayoutName string

// InitKeyboard initializes the keyboard layout from the build-time default.
func InitKeyboard() {
	currentLayoutName = initKeyboardLayout()
}

// SwitchLayout switches the active keyboard layout by name ("us" or "it").
// Interrupts are disabled during the swap to prevent IRQ1 from calling
// translateScancode while currentLayout is half-written (type ptr != data ptr).
// Must only be called with interrupts already enabled: it unconditionally calls
// EnableInterrupts() after the swap and does not preserve the prior interrupt state.
// Returns true if the layout was found and applied.
func SwitchLayout(name string) bool {
	var (
		keyLayout  layout.Layout
		layoutName string
	)

	switch name {
	case "us":
		keyLayout, layoutName = layout.GetUS()
	case "it":
		keyLayout, layoutName = layout.GetIT()
	default:
		return false
	}

	DisableInterrupts()
	keyboard.SetLayout(keyLayout)
	EnableInterrupts()
	currentLayoutName = layoutName
	return true
}

// GetCurrentLayoutName returns the name of the currently active layout.
func GetCurrentLayoutName() string {
	return currentLayoutName
}

var ticks uint64

func IRQ0Handler() {
	ticks++
	PICEOI(0)
	scheduler.Schedule()
}

func IRQ1Handler() {
	// Read & buffer scancode -> rune (no terminal printing here!)
	keyboard.IRQHandler()

	// Tell PIC we're done with IRQ1, otherwise it won't fire again
	PICEOI(1)
}

func GetTicks() uint64 {
	return ticks
}
