# DavOS Manual

This manual is designed to help people with limited operating systems background navigate `go-dav-os`.

Goal: understand how kernel components connect, and know where to look in the code when you want deeper detail.

## Where to start

1. Read `docs/manual/00-introduction/README.md`
2. Jump to the subsystem you care about (boot, memory, scheduler, filesystem, shell)

## Tree structure

```text
docs/manual/
├── README.md
├── STUDY_PATH.md
├── 00-introduction/
│   ├── README.md
│   └── glossary.md
├── 01-overview/
│   └── architecture.md
├── 02-boot/
│   ├── boot-and-grub.md
│   └── linker-and-initial-memory-layout.md
├── 03-kernel-core/
│   ├── main-loop.md
│   ├── interrupts-and-syscalls.md
│   ├── scheduler-and-tasks.md
│   └── theory-reference.md
├── 04-memory/
│   ├── multiboot-memory-map.md
│   └── page-frame-allocator.md
├── 05-io/
│   ├── vga-terminal.md
│   ├── ps2-keyboard.md
│   └── driver-ata-pio.md
├── 06-filesystem/
│   ├── filesystem-in-memory.md
│   └── fat16.md
├── 07-shell/
│   └── shell-commands.md
└── 08-build-and-test/
    ├── build-run.md
    └── test-debug.md
```

## Quick map: code -> docs

| Code | Topic | Manual section |
| --- | --- | --- |
| `boot/`, `iso/grub/` | Boot and long mode transition | `02-boot/` |
| `kernel/`, `asm/switch.s` | Kernel init, interrupts, scheduler | `03-kernel-core/` |
| `mem/` | Multiboot2 memory map and page allocator | `04-memory/` |
| `terminal/`, `keyboard/`, `drivers/ata/` | Basic I/O (video, input, disk) | `05-io/` |
| `fs/`, `fs/fat16/` | In-memory FS and persistent FAT16 | `06-filesystem/` |
| `shell/` | Command interface | `07-shell/` |
| `Makefile`, `Dockerfile`, `scripts/` | Build, run, test, debug | `08-build-and-test/` |

## Reading level

This manual starts from basics, then moves into real code. You do not need to understand everything at once: use each page as a map, then open the referenced source files for details.
