# User/Kernel Memory Isolation

This page explains how memory isolation is currently implemented between ring3 user code and ring0 kernel code in `go-dav-os`.

It focuses on the actual implementation in:

- `boot/boot.s` (early page table build)
- `boot/stubs_amd64.s` (user entry plumbing, #PF stub)
- `user/hello.s` (user payload and kernel-access probes)
- `kernel/task_runner.go` (program dispatch to user virtual addresses)
- `kernel/idt.go` (page-fault gate installation)
- `scripts/test_boot.py` (automated verification)

## 1. Security goal (current scope)

Current target:

- ring3 code can execute only from explicitly user-mapped pages
- ring3 reads/writes to kernel-mapped pages must fault
- kernel remains mapped in the same address space for kernel-mode execution

Out of scope (for now):

- per-process page tables
- process-to-process isolation
- demand paging / copy-on-write

## 2. Virtual address layout used today

The boot code defines a dedicated user virtual window:

- `USER_VA_BASE = 0x40000000`
- user window size: 8 KiB (`0x40000000 .. 0x40002000`)

Mapped pages:

- `0x40000000`: user program page (`.user_prog`)
- `0x40001000`: user stack page (RW)

Everything else in the 0..4 GiB identity map is kept supervisor-only.

## 3. How paging is built

In `setup_long_mode` (`boot/boot.s`):

1. Identity-map 0..4 GiB with 2 MiB pages (`pd0..pd3`).
2. Kernel identity mappings use flags `present|rw|ps` (`0x83`), so `U/S=0`.
3. Keep one user-capable walk path only where needed:
   - `pml4[0]` has `U/S=1`
   - `pdpt[1]` has `U/S=1` (contains user window)
4. Replace `pd1[0]` with a 4 KiB page table (`pt_user`).
5. Fill `pt_user` entries:
   - program page: `present|user` (`0x05`, read-only)
   - stack page: `present|rw|user` (`0x07`)

Result:

- kernel image and kernel data remain mapped and usable in ring0
- ring3 cannot access kernel identity pages because `U/S=0` on kernel mappings

## 4. User program mapping and launch path

`user/hello.s` places user payload in a page-aligned `.user_prog` section and exports:

- `go_0kernel.userHelloStart`
- `go_0kernel.userProbeReadKernelStart`
- `go_0kernel.userProbeWriteKernelStart`
- `__user_program_page`

`boot/stubs_amd64.s` computes user virtual entry points by:

- taking symbol offset from `__user_program_page`
- adding it to `USER_VA_BASE`

This keeps runtime entry addresses in the user VA window even though payload bytes are linked into the kernel image physically.

`kernel/task_runner.go` then dispatches:

- `run hello` -> normal user syscall demo
- `run kread` -> intentional ring3 read from kernel address
- `run kwrite` -> intentional ring3 write to kernel address

All are launched with `ExecuteUserTask(rip, rsp)` and `rsp = 0x40002000` (top of mapped user stack page).

## 5. Fault path for illegal user access

Illegal ring3 access to a supervisor page raises `#PF`.

Implementation:

- `boot/stubs_amd64.s` provides `PFaultStub` and emits `PF` on debug port `0xE9`
- `kernel/idt.go` installs vector `0x0E` with `getPFaultStubAddr()`

This gives an unambiguous marker in QEMU debug logs when isolation works as intended.

## 6. Automated verification

`scripts/test_boot.py` now includes dedicated probes:

- boot VM + run `kread` + expect `PF`
- boot VM + run `kwrite` + expect `PF`

Each probe runs in its own QEMU instance because fault handling is terminal in the current setup.

## 7. What this isolation guarantees (and what it does not)

Guaranteed now:

- ring3 cannot directly read/write kernel identity mappings
- kernel remains mapped and fully accessible in ring0
- user entry and user stack are explicit user pages

Not guaranteed yet:

- independent page tables per process
- isolated user address spaces between different tasks
- user-mode recovery from page faults (current #PF path halts)

## 8. Next hardening steps

Practical next steps if you want stronger isolation:

1. Allocate separate user page tables per task/process.
2. Add a recoverable user `#PF` path (kill task, keep kernel alive).
3. Move from static user pages to allocator-backed mappings.
4. Add syscall validation for user pointers against mapped user ranges.
