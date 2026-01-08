# go-dav-os

Hobby project to dig deeper into how an OS works by writing the bare essentials of a kernel in Go. Only the kernel lives here (gccgo, 32-bit protected mode); BIOS and bootloader are handled by battle-tested tools (GRUB with a Multiboot header). No reinvention of those pieces.

## What’s inside

- Boot: `boot/boot.s` exposes the Multiboot header and `_start`, sets up a 16 KB stack, disables interrupts, and jumps into `kernel.Main`
  - The Multiboot info pointer (GRUB `EBX`) is passed to `kernel.Main(...)` so the kernel can read boot-time memory info
  - A couple of freestanding helpers live in `boot/` as well (minimal stubs + `memcpy` to keep the build libc-free)

- Kernel: `kernel/` in Go, freestanding build with gccgo
  - IDT + PIC remap + PIT init
  - Tick counter from the PIT and a `hlt`-based idle loop when there’s no input

- Terminal: `terminal/` writes to VGA text mode 80x25, manages cursor, scroll, and backspace

- Keyboard: `keyboard/` reads from PS/2 and maps keys with the Italian layout only (temporary)

- Tiny shell: interactive prompt + basic line editing, commands are mostly for debugging

- Memory: `mem/`
  - Multiboot v1 memory map parsing (`mmap` command)
  - A minimal 4KB page frame allocator backed by a bitmap placed right after the kernel image (`pfa/alloc/free`)

- Filesystem: `fs/`
  - Minimal in-memory FS backed by allocated pages (`ls/write/cat/rm/stat`)

## Project status

- Experimental, single-core
- No paging, no scheduler, no real storage drivers yet
- Runs in 32-bit protected mode, meant for QEMU/GRUB, no UEFI
- Go runtime pared down: freestanding build (no standard library) with just the stubs the toolchain ends up expecting

## Dependencies

- Via Docker (recommended): Docker with `--platform=linux/amd64`
- Native (if you want to do it manually): cross toolchain `i686-elf-{binutils,gccgo}`, `grub-mkrescue`, `xorriso`, `mtools`, `qemu-system-i386`

## Build and run (Docker)

```bash
docker build --platform=linux/amd64 -t go-dav-os-toolchain .
docker run --rm --platform=linux/amd64 \
  -v "$PWD":/work -w /work go-dav-os-toolchain \
  make            # builds build/dav-go-os.iso
qemu-system-i386 -cdrom build/dav-go-os.iso
```

Quick targets from the Makefile
- `make docker-build-only` builds the image and the ISO
- `make run` (outside Docker) runs QEMU on an existing ISO

## Build natively

Assuming an `i686-elf-*` toolchain is installed

```bash
make
qemu-system-i386 -cdrom build/dav-go-os.iso
```

To force cross binaries: `make CROSS=i686-elf`

## What you’ll see on screen

- On boot the prompt `> ` shows up
- `help` lists commands, `clear` wipes the screen, `about` prints kernel info
- The kernel idles with `hlt` when nothing is happening

### Shell commands (current)

- `help`, `clear`, `about`, `echo`
- `ticks` (PIT tick counter)
- `mem <hex_addr> [len]` (hexdump)
- `mmap` (Multiboot memory map)
- `pfa`, `alloc`, `free <hex_addr>` (page allocator)
- `ls`, `write <name> <text...>`, `cat <name>`, `rm <name>`, `stat <name>` (filesystem)
- `version` (OS name and version)

## Other folder layout

- `iso/`: GRUB config and ISO packaging bits (grub.cfg)
- `boot/`: also contains the linker script (linker.ld) and a couple of freestanding helpers used by the build
- `build/`: build output (ISO + ELF)

## Contributing
Contributions are welcome 👊  
The project is still early-stage and intentionally minimal, so small PRs are the best way to help.

If you want to contribute, start here: **CONTRIBUTING.md**  
You can also check Issues labeled `good first issue` / `help wanted`, or open a Discussion if you want to propose something.

## Final note

Personal, open-source, work-in-progress. I’m building pieces as I learn them—the goal is understanding, not chasing modern-OS feature lists
