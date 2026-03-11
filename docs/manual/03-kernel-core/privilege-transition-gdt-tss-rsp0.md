# GDT, TSS, and `RSP0` in `go-dav-os`

This document explains exactly what the Global Descriptor Table (GDT), Task State Segment (TSS), and `RSP0` are, why they matter, and how they are implemented in this codebase.

It covers both:
- Go code (`kernel/gdt.go`, `kernel/tss.go`, `kernel/gdt/gdt.go`, `kernel/tss/tss.go`, `kernel/kernel.go`, `kernel/idt.go`)
- Assembly stub (`boot/stubs_amd64.s`)

## 1. Why this exists

On x86_64, an interrupt/trap coming from user mode (CPL3) into kernel mode (CPL0) must switch to a trusted kernel stack.

Without this, the CPU would keep using the user stack during kernel entry, which is unsafe and can corrupt privilege boundaries.

The CPU performs this stack switch automatically only if:
- a valid TSS is loaded in `TR` (`ltr`), and
- `TSS.RSP0` points to a valid kernel stack top.

This implementation provides exactly that foundation.

## 2. Quick theory

## 2.1 GDT in long mode

In 64-bit mode, segmentation is mostly disabled for linear address translation, but segment descriptors still matter for:
- privilege levels (ring0 vs ring3)
- code/data selectors used by `CS`, `SS`, etc.
- system descriptors, especially the TSS descriptor

So even in long mode, we still need a valid GDT.

## 2.2 TSS in long mode

In 64-bit mode, the TSS is **not** used for hardware task switching (old x86 feature). Instead, it is used mainly for:
- kernel stack pointers for privilege transitions (`RSP0`, `RSP1`, `RSP2`)
- optional Interrupt Stack Table (IST)

For user→kernel entry, `RSP0` is the key field.

## 2.3 What `RSP0` is

`RSP0` is the stack pointer the CPU loads when entering ring0 from a less privileged ring (typically ring3), via an interrupt/trap gate.

When such a transition happens, hardware does this (simplified):
1. Read `RSP0` from current TSS
2. Switch to that stack
3. Push old user context (`SS`, `RSP`, `RFLAGS`, `CS`, `RIP`)
4. Transfer control to the handler entry

That means your handler executes on a known-good kernel stack.

## 3. GDT/TSS layout used here

Defined in `kernel/gdt.go` + `kernel/tss.go`:

- Selector `0x00`: null descriptor
- Selector `0x08`: kernel code segment (`kernelCodeSelector`)
- Selector `0x10`: kernel data segment (`kernelDataSelector`)
- Selector `0x1B`: user code segment (`userCodeSelector`, RPL=3)
- Selector `0x23`: user data segment (`userDataSelector`, RPL=3)
- Selector `0x28`: TSS descriptor (`tssSelector`), 16-byte system descriptor across two GDT entries

Important detail: in long mode, a TSS descriptor consumes **two** GDT slots (128 bits).

## 4. Go implementation details

## 4.1 Files: `kernel/gdt.go` and `kernel/tss.go`

### Constants and selectors

The file defines:
- segment selectors (`kernelCodeSelector`, `kernelDataSelector`, `userCodeSelector`, `userDataSelector`, `tssSelector`)
- descriptor bit patterns for kernel/user code+data descriptors

### Exact TSS memory image

`cpuTSS` is stored as a raw `[104]byte` (`tss.TSSSize = 104`) instead of a Go struct.

Reason:
- Go struct padding/alignment can change offsets
- x86_64 TSS requires exact byte offsets

The implementation writes fields with explicit helpers in `kernel/tss/tss.go`:
- `SetRSP0(...)` for `RSP0` at offset 4
- `SetIomapBase(...)` for I/O bitmap base at offset 102

This guarantees the CPU reads the expected layout.

### Kernel trap stack

A dedicated static stack is allocated:
- `trapStack [4096]byte`

`defaultKernelTrapStackTop()` computes top-of-stack and aligns it to 16 bytes.

`SetKernelRSP0()` programs TSS `RSP0` with this top address.

### Building and loading GDT+TSS

`InitGDTAndTSS()` performs the full sequence:
1. Fill GDT entries 0..4 (null, kernel/user segments)
2. Program TSS fields (`IomapBase`, `RSP0`)
3. Encode TSS descriptor into GDT entries 5 and 6 (`tss.EncodeTSSDescriptor`)
4. Build GDTR (`gdt.PackGDTR`)
5. Load GDT (`LoadGDT` assembly helper)
6. Reload data segments (`LoadDataSegments`)
7. Load task register (`LoadTR(tssSelector)`)

After step 7, `TR` points to a valid TSS and ring3→ring0 stack switching can work.

## 4.2 File: `kernel/kernel.go`

`Main()` now calls:
- `InitGDTAndTSS()` **before** `InitIDT()`

Why ordering matters:
- IDT entries can be active only after interrupt setup
- if user-mode transitions happen, TSS must already be valid

## 4.3 File: `kernel/idt.go`

IDT setup now uses `kernelCodeSelector` explicitly for handlers (including `int 0x80` gate target selector), instead of reading current `CS` dynamically.

This makes IDT descriptor selectors explicit and consistent with the runtime GDT.

`int 0x80` still has DPL=3 (`0xEE`) so user-space can invoke it.

## 5. Assembly implementation details

## File: `boot/stubs_amd64.s`

Three new helper entry points were added:

- `go_0kernel.LoadGDT`
  - Executes `lgdt (%rdi)`
  - `RDI` points to packed 10-byte GDTR

- `go_0kernel.LoadTR`
  - Moves selector from `DI` to `AX`
  - Executes `ltr %ax`

- `go_0kernel.LoadDataSegments`
  - Loads `DS/ES/SS/FS/GS` from selector in `DI`

These are minimal privileged instructions exposed to Go.

## 6. End-to-end flow (intended ring3 path)

When a real ring3 context executes `int 0x80`:

1. CPU sees IDT gate with DPL allowing ring3 call
2. Privilege change CPL3→CPL0 is required
3. CPU reads current TSS via `TR`
4. CPU loads kernel stack from `TSS.RSP0`
5. CPU pushes user return frame on new kernel stack
6. Entry stub runs (`Int80Stub`), saves registers, calls `Int80Handler`
7. Handler returns, stub executes `iretq`
8. CPU restores user frame and returns to ring3

The crucial safety property is step 4: handler logic runs on a trusted kernel stack, not user-controlled memory.

## 7. Current scope and limitation in this repo

This work enables safe stack switching for user→kernel transitions, but it is only the foundation.

At the moment, the sample task (`user/hello.s`) is linked into the kernel image and scheduled like a normal kernel task unless a full ring3 launcher is added (with `iretq` to user selectors and user stack).

So:
- GDT/TSS/RSP0 infrastructure is now correct and ready
- full user-mode process entry is a separate next step

## 8. Why this design is robust

- Uses explicit selectors and descriptors instead of implicit bootstrap state
- Uses raw-byte TSS encoding to avoid layout bugs
- Loads `TR` explicitly with a valid 64-bit TSS descriptor
- Keeps `RSP0` programmable via `SetKernelRSP0()` for future per-task kernel stacks
