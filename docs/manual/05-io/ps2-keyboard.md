# PS/2 Keyboard

## What it does

Reads PS/2 scancodes, converts them to runes (Italian layout), and pushes them into a ring buffer.

Key files:

- `keyboard/keyboard.go`
- `keyboard/irq.go`
- `keyboard/layout.go`

## Operating modes

- Blocking: `ReadKey()` (not used in the current main kernel loop)
- Non-blocking: `TryRead()` on ring buffer

## Pipeline currently used

1. IRQ1 calls `keyboard.IRQHandler()`
2. handler reads from port `0x60`
3. release scancodes are ignored (high bit set)
4. scancode is mapped via `LayoutIT`
5. rune is pushed into ring buffer
6. main loop consumes with `TryRead()`

## Current limitations

- minimal Italian layout only
- no advanced Shift/Ctrl/Alt state handling
- if buffer is full, new characters are dropped
