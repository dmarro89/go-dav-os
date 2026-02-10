# Multiboot Memory Map

## What it does

Parses the Multiboot2 structure from GRUB and stores a compact local copy of memory regions.

Key file:

- `mem/multiboot.go`

## Flow

`InitMultiboot(mbInfoAddr)`:

1. validates pointer and `totalSize`
2. iterates Multiboot2 tags aligned to 8 bytes
3. finds tag type `6` (memory map)
4. copies entries into `mmapEntries` (max 64)

Each stored entry contains:

- base physical address (hi/lo)
- length (hi/lo)
- type (`1` means usable RAM)

## API used by other modules

- `MMapCount()`
- `MMapEntry(i)`

Shell commands `mmap` and `mmapmax` use these APIs directly.
