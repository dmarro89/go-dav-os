# Page Frame Allocator (PFA)

## What it does

Allocates and frees 4 KiB physical pages using a bitmap.

Key file:

- `mem/allocator.go`

## Strategy

1. compute highest usable end address from type-1 regions
2. derive `totalPages`
3. place bitmap in usable memory after kernel/bootstrap areas
4. mark all pages as used
5. mark type-1 pages as free
6. reserve low memory + kernel + bitmap pages as used

## Main API

- `InitPFA()`
- `PFAReady()`
- `AllocPage()` -> physical address or `0`
- `FreePage(addr)` -> `bool`
- `TotalPages()`, `FreePages()`, `UsedPages()`

## Current limitations

- works on identity-mapped physical memory
- no process virtual memory support
- no ownership protection for allocated pages

## Useful shell commands

- `pfa`
- `alloc`
- `free <hex_addr>`
