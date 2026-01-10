# x86_64 Migration Plan

This document outlines the steps required to migrate `go-dav-os` from 32-bit protected mode to 64-bit long mode (x86_64).

## 1. Bootloader Changes (Long Mode Setup)
Current state: Boots via Multiboot 1 (32-bit protected mode).
Goal: Reach 64-bit Long Mode.

**Note**: Multiboot v1 is 32-bit only. We should switch to **Multiboot2** to support modern 64-bit loading properly, or use a trampoline. This plan assumes moving to Multiboot2.

### Steps:
1.  **Paging Setup**:
    *   Create a Page Map Level 4 (PML4) table.
    *   Create a Page Directory Pointer Table (PDPT).
    *   Create a Page Directory (PD).
    *   Identity map the first 1GB (or enough for kernel + boot structures) using huge pages (2MB) for simplicity.
    *   Update `boot.s` header to use **Multiboot2** tags.
    *   Load clean CR3 with the address of PML4.

2.  **Enable Long Mode**:
    *   Set PAE bit in CR4.
    *   Set LME (Long Mode Enable) bit in EFER MSR (Model Specific Register `0xC0000080`).
    *   Enable Paging (PG bit in CR0).

3.  **GDT Update**:
    *   Define a 64-bit GDT (Global Descriptor Table). The code segment descriptor must have the 'L' bit set (Long mode).
    *   Load the new GDT (`lgdt`).

4.  **Jump**:
    *   Perform a long jump to the 64-bit code segment to update `CS`.

## 2. Kernel Changes (Go Runtime)
Current state: `GOARCH=386`
Goal: `GOARCH=amd64`

### Steps:
1.  **Compiler Target**:
    *   Update `Makefile` to set `GOARCH=amd64` for `gccgo`.
    *   Ensure cross-compiler `x86_64-elf-gcc` / `x86_64-elf-gccgo` is used.
    *   Update `linker.ld` to output `elf64-x86-64` and handle 64-bit sections.

2.  **Entry Point**:
    *   Update `boot.s` to pass Multiboot info pointer in a 64-bit register (e.g., `rdi` for C ABI or stack depending on Go ABI).
    *   Update `kernel.Main` signature in `kernel/kernel.go`:
        ```go
        // from func Main(addr uint32)
        func Main(addr uint64) // or uintptr
        ```

3.  **Memory Management (`mem/`)**:
    *   Refactor `allocator.go` to use `uint64` or `uintptr` for physical addresses.
    *   The current bitmap allocator assumes 32-bit addresses. It needs to handle >4GB memory ranges if available from Multiboot memory map.
    *   Struct `mmapEntry` in `mem/multiboot.go` may ideally remain compatible if we parse the raw bytes correctly, but internal representation should standardise on 64-bit types.

4.  **Interrupts (`kernel/idt.go`, `kernel/pic.go`)**:
    *   **IDT**: In 64-bit mode, IDT entries are 16 bytes (was 8 bytes). This requires a new struct definition and loader logic.
    *   **Handlers**: Assembly interrupt stubs (`boot/boot.s`) must use `iretq` (64-bit return) and save/restore 64-bit registers (`rax`, `rbx`... `r15`).

5.  **Syscalls**:
    *   (Future Work): Once the kernel is stable in 64-bit, switch system calls to use `syscall`/`sysret` instructions instead of `int 0x80`.
    *   This will require updating the userspace/kernel boundary ABI.

## 3. Architecture Independent Parts (No Change Needed)
The following components are largely high-level and should remain mostly untouched, assuming `int` / `uint` behave as expected (Go `int` becomes 64-bit on amd64):

*   **Filesystem (`fs/`)**: FAT driver logic is independent of CPU bit-mode.
*   **Shell (`shell/`)**: Command parsing and logic are generic.
*   **Keyboard (`keyboard/`)**: Logic for scancodes is the same, though port I/O wrappers `inb`/`outb` assembly needs 64-bit register usage updates (e.g. `mov dx, [rsp+...]`).
*   **Terminal (`terminal/`)**: VGA buffer manipulation (0xB8000) is standard, though pointer arithmetic will be 64-bit.

## 4. Minimal Roadmap Checklist

- [ ] **Phase 1: Toolchain & Build**
    *   [ ] Update `Makefile` variables for `x86_64` toolchain.
    *   [ ] Update `Dockerfile` to include `x86_64-elf-gcc` and `qemu-system-x86_64`.
    *   [ ] Update `linker.ld` for 64-bit output.

- [ ] **Phase 2: Bootloader**
    - [ ] Modify `boot.s` to implement paging and long mode switch.
    - [ ] Stub out the Go kernel entry for a simple "Hello 64-bit" asm check.

- [ ] **Phase 3: Kernel Port**
    - [ ] Update `kernel.Main` and `allocator` to 64-bit.
    - [ ] Fix assembly stubs (`inb`, `outb`, `lidt`, `lgdt`) for 64-bit registers.
    - [ ] Get "Hello World" from Go kernel on screen.

- [ ] **Phase 4: Interrupts & Input**
    - [ ] Implement 64-bit IDT.
    - [ ] Port PIT and Keyboard drivers.
    - [ ] Verify Shell works.
