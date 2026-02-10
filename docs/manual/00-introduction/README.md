# Introduction

## What DavOS is

`go-dav-os` is a minimal kernel written in Go (compiled with `gccgo`) with critical low-level parts in x86_64 assembly.

It is not trying to be a full production OS. It is a learning project focused on core operating system mechanics.

## What is in this project

- Boot with GRUB + Multiboot2
- x86_64 long mode transition
- Kernel with basic interrupts (timer and keyboard)
- Very small round-robin scheduler
- Physical memory management using 4 KiB pages
- In-memory filesystem
- ATA PIO disk driver + FAT16
- Text shell with diagnostic commands

## What is not in this project (yet)

- Advanced process isolation
- Full per-process virtual memory
- Multi-core support
- Modern drivers (AHCI, USB, networking, and so on)

## Minimal prerequisites

You do not need to be an OS expert. It helps if you know:

- basic Assembler/Go
- what an interrupt handler is (high-level idea)
- difference between RAM and disk

If terms are unfamiliar, start with the glossary first.
