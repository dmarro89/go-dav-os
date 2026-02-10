# VGA Terminal

## What it does

Implements text output on VGA 80x25 text mode and updates the hardware cursor.

Key file:

- `terminal/terminal.go`

## Core concepts

- Video memory base: `0xB8000`
- Each cell: character byte + color attribute byte
- Cursor updates through ports `0x3D4` and `0x3D5`

## Main functions

- `Init()`, `Clear()`
- `PutRune()`, `Print()`
- `Backspace()`
- `PrintInt()`

## Practical notes

- Supports newline, backspace, and vertical scroll
- Also mirrors output to debug port `0xE9` (`debugChar`) for headless QEMU debugging
