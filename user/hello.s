# hello.s - Tiny userland program for go-dav-os
# Validates syscall interface: SYS_WRITE and SYS_EXIT

.code64
.section .text
.global _start

_start:
    # SYS_WRITE: write "hello from userland\n" to stdout
    mov  $1, %rax           # SYS_WRITE = 1
    mov  $1, %rbx           # fd = 1 (stdout)
    lea  msg(%rip), %rcx    # buffer address
    mov  $21, %rdx          # length = 21 bytes
    int  $0x80              # invoke syscall

    # SYS_EXIT: terminate with status 0
    mov  $2, %rax           # SYS_EXIT = 2
    xor  %rbx, %rbx         # status = 0
    int  $0x80              # invoke syscall

    # Should never reach here, but just in case:
    hlt

.section .rodata
msg:
    .ascii "hello from userland\n"
