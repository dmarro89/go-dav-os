# Syscall Entry And ABI

This page explains the current x86_64 syscall path in `go-dav-os`.

It covers:

- how user code enters the kernel with `syscall`
- how the kernel configures `STAR`, `LSTAR`, `SFMASK`, and `EFER.SCE`
- the current syscall ABI
- why the return path currently uses `iretq` instead of `sysretq`

Code layout for this subsystem is now split as follows:

- `boot/stubs_amd64.s`: low-level syscall entry/return assembly and MSR helpers
- `kernel/syscall/`: Go-side ABI types, MSR setup logic, and syscall dispatcher
- `kernel/syscall_bridge.go`: thin bridge between assembly entrypoints and the Go package

## 1. What changed

User-mode programs no longer use `int 0x80` as the primary syscall path.

Instead:

- user code executes `syscall`
- CPU jumps to the kernel entrypoint from `LSTAR`
- the assembly entry stub builds a 64-bit trapframe
- Go dispatches the syscall through `kernel/syscall`
- the stub returns to ring3 with `iretq`

The older `int 0x80` path is still present as a compatibility/fallback path.

## 2. Why `syscall` is not the same as `int 0x80`

Important architectural difference:

- `int 0x80` can switch to `TSS.RSP0` automatically on a ring3 -> ring0 transition
- `syscall` does **not** switch to a kernel stack automatically

Because of that, the kernel must switch stacks explicitly in the syscall entry stub before calling Go code.

## 3. MSR configuration

The runtime syscall setup writes these MSRs:

- `IA32_STAR` (`0xC0000081`)
- `IA32_LSTAR` (`0xC0000082`)
- `IA32_SFMASK` (`0xC0000084`)
- `IA32_EFER` (`0xC0000080`) with `SCE=1`

In this repo, the Go-side setup lives in `kernel/syscall/runtime.go`.

Current policy:

- `STAR` selects the kernel code segment for entry and the user code segment for future syscall/sysret pairing
- `LSTAR` points to the kernel syscall entry stub
- `SFMASK` currently clears `IF` on entry
- `EFER.SCE` enables the `syscall` instruction family

## 4. Current syscall ABI

Current ABI is:

- `RAX`: syscall number
- `RDI`, `RSI`, `RDX`, `R10`, `R8`, `R9`: arguments
- `RAX`: return value
- `RCX`, `R11`: clobbered by CPU on entry

Implemented syscalls:

- `SYS_WRITE`
  - `RDI = fd`
  - `RSI = buf`
  - `RDX = len`
- `SYS_EXIT`
  - `RDI = status`
- `SYS_GETTICKS`
  - no arguments

The syscall numbers and `TrapFrame` type are defined in `kernel/syscall/abi.go`.

## 5. Trapframe shape

The syscall entry stub synthesizes the same general trapframe layout used by the interrupt-gate path:

- general registers
- saved user `RIP`
- saved `CS`
- saved user `RFLAGS`
- saved user `RSP`
- saved `SS`

This keeps the Go dispatcher independent from whether the call came from `int 0x80` or `syscall`.

The shared dispatcher itself lives in `kernel/syscall/dispatch.go`.

## 6. Why return uses `iretq` today

Although entry now uses `syscall`, the return path still uses `iretq`.

Reason:

- `iretq` is easier to integrate with the shared trapframe layout
- it avoids early `sysretq` corner cases while the ABI and stack-switch logic are still minimal
- it works cleanly with the current `SYS_EXIT` path, which can unwind directly back to the kernel launcher

So the current model is:

- entry via `syscall`
- return via `iretq`

That is intentional for simplicity and robustness in this stage of the kernel.

## 7. Validation

Current checks:

- unit tests cover syscall MSR packing and Go-side dispatcher behavior
- boot tests verify `run hello` prints through `SYS_WRITE` and returns through `SYS_EXIT`
- protection-fault probes (`run kread`, `run kwrite`) still verify user/kernel isolation

## 8. Next step if needed

The natural next refinement is a true `sysretq` return path.

Before doing that, the kernel should first:

1. finalize selector assumptions for `STAR`
2. harden the syscall entry stack strategy
3. verify return semantics for user `RCX`/`R11` and flags masking
