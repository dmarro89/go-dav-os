# Minimal Glossary

## Bootloader
Program that loads the kernel into memory and transfers control to its entrypoint.

## Multiboot2
Specification used by GRUB to pass boot information to the kernel (for example memory map).

## Long mode
64-bit execution mode on x86_64 CPUs.

## IDT (Interrupt Descriptor Table)
Table telling the CPU which handler to call for each interrupt/exception.

## IRQ
Hardware interrupt (for example timer or keyboard).

## PIC / PIT
- PIC: legacy interrupt controller.
- PIT: periodic hardware timer.

## Syscall
Controlled request from task/user code to kernel services.

## Page frame
Physical memory block (4 KiB in this project).

## PFA (Page Frame Allocator)
Allocator that assigns and frees physical page frames.

## ATA PIO
Simple disk I/O mode using I/O ports.

## FAT16
Classic on-disk filesystem, used here in a minimal implementation.
