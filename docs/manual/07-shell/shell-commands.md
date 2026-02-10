# Shell and Commands

## What it does

The shell provides the text prompt and dispatches commands to kernel, memory, filesystem, and drivers.

Key file:

- `shell/shell.go`

## Input behavior

- fixed prompt `> `
- newline and backspace handling
- history of last 32 commands (`history`)
- typo suggestions (string distance)

## Command groups

System:

- `help`
- `clear`
- `echo <text>`
- `version`
- `ticks`
- `history`

Memory:

- `mem <hex_addr> [len]`
- `mmap`
- `mmapmax`
- `pfa`
- `alloc`
- `free <hex_addr>`

RAM filesystem:

- `ls`
- `write <name> <text...>`
- `cat <name>`
- `rm <name>`
- `stat <name>`

Disk/FAT16:

- `disk read <lba>`
- `disk write <lba> <text>`
- `fatformat`
- `fatinit`
- `fatinfo`
- `fatls`
- `fatcreate <name> <content>`
- `fatread <name>`

Tasks:

- `run <program>` (currently `hello`)

## Kernel wiring

`kernel.Main` connects shell dependencies through:

- `shell.SetTickProvider(GetTicks)`
- `shell.SetProgramRunner(RunProgram)`
