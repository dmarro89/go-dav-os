# go-dav-os

Hobby project to dig deeper into how an OS works by writing the bare essentials of a kernel in Go. Only the kernel lives here (gccgo, x86_64 long mode); BIOS and bootloader are handled by battle-tested tools (GRUB with a Multiboot2 header). No reinvention of those pieces.
Join [Discord](https://discord.gg/mBHhPZ65eW) for real time discussions on the project ! 

## What’s inside

- Boot: `boot/boot.s` exposes the Multiboot2 header and `_start`, sets up a 16 KB stack, enables long mode, and jumps into `kernel.Main`
  - The Multiboot2 info pointer is passed to `kernel.Main(...)` in `RDI`
  - Freestanding helpers live in `boot/` as well (minimal stubs + `memcmp` to keep the build libc-free)

- Kernel: `kernel/` in Go, freestanding build with gccgo
  - IDT + PIC remap + PIT init
  - Tick counter from the PIT and a `hlt`-based idle loop when there’s no input

- Terminal: `terminal/` writes to VGA text mode 80x25, manages cursor, scroll, and backspace

- Keyboard: `keyboard/` reads from PS/2 and maps keys with the Italian layout only (temporary)

- Tiny shell: interactive prompt + basic line editing, commands are mostly for debugging

- Memory: `mem/`
  - Multiboot2 memory map parsing (`mmap` and `mmapmax` commands)
  - A minimal 4KB page frame allocator backed by a bitmap placed inside usable memory (`pfa/alloc/free`)

- Filesystem: `fs/`
  - Minimal in-memory FS backed by allocated pages (`ls/write/cat/rm/stat`)

- Persistent Storage: `drivers/ata` + `fs/fat16`
  - ATA PIO driver for disk I/O
  - FAT16 filesystem with file create/read/list operations
  - Data persists across reboots on a 20MB disk image
  
## Architecture

![DavOS Architecture](docs/architecture.png)

## Project status

- Experimental, single-core
- 64-bit only (x86_64 long mode); 32-bit is no longer supported
- Basic paging (identity map), FAT16 persistent storage driver
- Runs in x86_64 long mode, meant for QEMU/GRUB, no UEFI
- Go runtime pared down: freestanding build (no standard library) with just the stubs the toolchain ends up expecting

## Dependencies

- Via Docker (recommended): Docker with `--platform=linux/amd64`
- Native (if you want to do it manually): cross toolchain `x86_64-elf-{binutils,gccgo}`, `grub-mkrescue`, `xorriso`, `mtools`, `qemu-system-x86_64`

## Build and run (Docker)

```bash
docker build --platform=linux/amd64 -t go-dav-os-toolchain .
docker run --rm --platform=linux/amd64 \
  -v "$PWD":/work -w /work go-dav-os-toolchain \
  make            # builds build/dav-go-os.iso
qemu-system-x86_64 -cdrom build/dav-go-os.iso
```

Quick targets from the Makefile
- `make docker-build-only` builds the image and the ISO
- `make run` (outside Docker) runs QEMU on an existing ISO

## Build natively

Assuming an `x86_64-elf-*` toolchain is installed

```bash
make
qemu-system-x86_64 -cdrom build/dav-go-os.iso
```

To force cross binaries: `make CROSS=x86_64-elf`

## What you’ll see on screen

- On boot the prompt `> ` shows up
- `help` lists commands, `clear` wipes the screen
- The kernel idles with `hlt` when nothing is happening

### Shell commands (current)

- `help`, `clear`, `echo`, `version`, `history`
- `ticks` (PIT tick counter)
- `mem <hex_addr> [len]` (hexdump)
- `mmap`, `mmapmax` (Multiboot memory map and highest usable end)
- `pfa`, `alloc`, `free <hex_addr>` (page allocator)
- `ls`, `write <name> <text...>`, `cat <name>`, `rm <name>`, `stat <name>` (in-memory filesystem)
- `run <program>` (task runner)

### Persistent Storage (FAT16)

- `fatformat` - Initialize disk with FAT16 structure
- `fatinit` - Mount the filesystem
- `fatinfo` - Show filesystem layout
- `fatls` - List files in root directory
- `fatcreate <name> <content>` - Create a file
- `fatread <name>` - Read a file  
- `disk read|write <lba>` - Raw sector access

**Example:**
```bash
fatformat
fatinit
fatcreate hello Hello World
fatls
fatread hello
```

### Run simple programs

The task runner can start small assembly programs linked into the kernel.
At the moment, the supported launch name is:
- `run hello`

This starts `user/hello.s`, which calls `SYS_WRITE` and then `SYS_EXIT`.

**Example:**
```bash
run hello
```

## Other folder layout

- `iso/`: GRUB config and ISO packaging bits (grub.cfg)
- `boot/`: also contains the linker script (linker.ld) and a couple of freestanding helpers used by the build
- `build/`: build output (ISO + ELF)

## Contributing

Contributions are welcome! This project is still early-stage and intentionally minimal, so **small PRs** are the best way to help.

- Review [CONTRIBUTING.md](CONTRIBUTING.md) for guidelines
- Check issues labeled [`good first issue`](../../labels/good%20first%20issue) or [`help wanted`](../../labels/help%20wanted)
- Open a [Discussion](../../discussions) to propose an idea or ask a question
  

## Final note

Personal, open-source, work-in-progress. I’m building pieces as I learn them—the goal is understanding, not chasing modern-OS feature lists

## Contribution
Thanks to:
[@metacatdud](https://github.com/metacatdud)for taking care of the entire migration from 32 to 64 bit architecture - really amazing work!<br/>
[@ranjan42](https://github.com/ranjan42)for the useful documentation added and the implementation of the scheduled and the command history<br/>
[@jgafnea](https://github.com/jgafnea) for improving the shell and documenting the contributing section<br/>
[@soorya38](https://github.com/soorya38)for taking care of a missing part of the project - the unit tests!<br/>
