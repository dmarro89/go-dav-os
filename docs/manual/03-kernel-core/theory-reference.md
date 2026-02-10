# Kernel Core Theory Reference

## Why this page exists

Chapter 3 mixes many low-level mechanisms (IDT, IRQ, PIC, PIT, syscalls, context switches).
This page is the theory companion: it explains the concepts independently from implementation details.

## 1. Interrupts, exceptions, and traps

On x86_64, control can be transferred to privileged handlers through different event classes:

- `Exception`: CPU-detected condition while executing instructions (for example `#GP`, `#DF`)
- `Interrupt`: asynchronous hardware event (for example timer, keyboard)
- `Software interrupt/trap`: explicit instruction (`int 0x80`) used here for syscalls

All of these use an IDT vector to choose the handler entry.

## 2. IDT basics

The Interrupt Descriptor Table (IDT) is an array of gate descriptors indexed by vector number.

- vectors are in range `0..255`
- on x86_64, each IDT entry is 16 bytes
- CPU uses `IDTR` register to find the table base and limit

Important gate fields (conceptually):

- target handler address
- code segment selector
- type/flags
- privilege level (`DPL`)

In this project, vector `0x80` is configured with user-callable privilege (`DPL=3`) for syscalls.

## 3. IRQ and PIC model

Hardware devices signal interrupts as IRQ lines. In legacy x86, IRQ routing is managed by the 8259 PIC pair:

- Master PIC handles IRQ `0..7`
- Slave PIC handles IRQ `8..15` through master's IRQ2 cascade

Why remapping is needed:

- default PIC vector offsets overlap CPU exception vectors
- remapping to `0x20`/`0x28` avoids that overlap

Why EOI is required:

- after servicing an IRQ, kernel must send End Of Interrupt
- without EOI, same IRQ line may stop firing

## 4. PIT timer model

The PIT is a periodic hardware timer used to generate IRQ0.

Core relation:

- PIT input frequency is `1193182 Hz`
- divisor = `1193182 / target_hz`

In this kernel, `target_hz = 100`, so periodic timer interrupts drive time ticks and scheduler entry points.

## 5. Syscall model in this project

This kernel uses an educational syscall path based on `int 0x80`.

High-level flow:

1. task executes software interrupt
2. CPU dispatches vector `0x80`
3. assembly stub saves context
4. Go handler reads syscall number/arguments from saved registers
5. return value is written back to return register

This is a custom ABI, intentionally minimal and not Linux-compatible.

## 6. Trap frame concept

A trap frame is the saved CPU context at interrupt/syscall entry.
It allows the handler to:

- inspect caller registers
- dispatch by syscall/event type
- modify return register values
- resume execution consistently

In this project, `TrapFrame` mirrors the register save order used by assembly stubs.

## 7. Critical sections and CPU idling

Two common low-level patterns used in chapter 3:

- `cli`/`sti`: temporarily disable/enable interrupts around small critical sections
- `hlt`: idle until next interrupt, reducing busy polling

Used correctly, this gives deterministic short critical regions and low idle CPU overhead.

## 8. Context switch theory

A context switch moves CPU execution from one task to another by:

1. saving current task register/stack pointer state
2. loading next task saved stack pointer
3. restoring registers and returning into the next task's control flow

The scheduler policy (round-robin here) decides *which* task runs next; the switch routine handles *how* CPU state is transferred.

## 9. Suggested external reading

- Intel 64 and IA-32 Architectures Software Developer's Manual (Vol. 3A):
  - Interrupt/exception architecture
  - IDT and gate descriptors
  - privilege levels and interrupt return
- OSDev Wiki topics:
  - Interrupt Descriptor Table
  - 8259 PIC
  - Programmable Interval Timer
  - System Calls

Use these references for architecture-level details beyond the scope of this project manual.
