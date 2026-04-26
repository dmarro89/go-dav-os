package kernel

import (
	"github.com/dmarro89/go-dav-os/keyboard"
	"github.com/dmarro89/go-dav-os/keyboard/layout"
)

func initKeyboardLayout() string {
	keyLayout, layoutName := layout.GetIT()
	keyboard.SetLayout(keyLayout)
	return layoutName
}
