/* boot/boot.s
 * Kernel entry point + Multiboot header for GRUB.
 *
 * Flow (Multiboot context):
 * - Provide the Multiboot2 header so GRUB recognizes and loads this image.
 * - GRUB jumps to _start with EAX=0x36D76289 and EBX pointing to the multiboot
 *   info struct (per spec).
 * - We immediately disable interrupts (cli) because no IDT/PIC is set up yet.
 * - We set ESP to a known 16 KB stack in .bss, aligned to 16 bytes.
 * - We park the CPU in a HLT loop as a placeholder until real kernel init runs.
 */
.code32

/* ---------------------------
 * Multiboot2 header section
 * ---------------------------
 * Placed in its own section so GRUB can locate it in the first 32 KiB.
 */
.set MULTIBOOT_MAGIC, 0xE85250D6
.set MULTIBOOT_ARCH,  0

.section .multiboot2
.align 8

multiboot_header_start:
	.long MULTIBOOT_MAGIC
	.long MULTIBOOT_ARCH
	.long multiboot_header_end - multiboot_header_start
	.long -(MULTIBOOT_MAGIC + MULTIBOOT_ARCH + (multiboot_header_end - multiboot_header_start))

# Info request tag: memory map (6), ELF sections (9)
	.align 8
	.word 1
	.word 0
	.long 16
	.long 6
	.long 9

# Entry address tag
	.align 8
	.word 3
	.word 0
	.long 16
	.long _start
	.long 0

/* End tag */
	.align 8
	.word 0
	.word 0
	.long 8

multiboot_header_end:

// --- Stack ---
.section .bootstrap_stack, "aw", @nobits

.align 16
stack_bottom:
	.skip 16384              # 16 KB di stack
stack_top:

# Long mode paging structures (4 KiB aligned, zero-initialized).
.align 4096
pml4:
	.skip 4096
.align 4096
pdpt:
	.skip 4096
.align 4096
pd:
	.skip 4096

.align 4
multiboot_info_ptr:
	.long 0

.section .rodata
.align 8
gdt64:
	.quad 0x0000000000000000
	.quad 0x00AF9A000000FFFF
	.quad 0x00CF92000000FFFF

gdt64_desc:
	.word (gdt64_end - gdt64 - 1)
	.long gdt64

gdt64_end:

/* ---------------------------
 * Executable code
 * ---------------------------
 * GRUB jumps here after validating the header, with:
 * - EAX = 0x2BADB002 (Multiboot magic passed to the kernel)
 * - EBX = pointer to the Multiboot info structure
 */
	.section .text
	.global  _start
	.type    _start, @function

_start:
	cli # disable interrupts (no IDT/PIC set yet)

# initialize ESP to the top of our 16 KB stack in .bss
	mov  $stack_top, %esp

# Multiboot2: EBX contains the address of the multiboot info structure.
	movl %ebx, multiboot_info_ptr

	call setup_long_mode

	cli

.Lhang:
	jmp .Lhang
.size _start, . - _start

setup_long_mode:
# Build minimal identity-mapped paging (1 GiB via 2 MiB pages).
	lea pml4, %edi
	movl $pdpt, %eax
	orl $0x03, %eax
	movl %eax, (%edi)
	movl $0, 4(%edi)

	lea pdpt, %edi
	movl $pd, %eax
	orl $0x03, %eax
	movl %eax, (%edi)
	movl $0, 4(%edi)

	lea pd, %edi
	xorl %ecx, %ecx
	movl $0x83, %edx           # present|rw|ps

.Lmap_2m:
	movl %ecx, %eax
	shll $21, %eax             # ecx * 2 MiB
	orl %edx, %eax
	movl %eax, (%edi)
	movl $0, 4(%edi)
	addl $8, %edi
	incl %ecx
	cmpl $512, %ecx
	jne .Lmap_2m

# Load PML4 and enable PAE.
	movl $pml4, %eax
	movl %eax, %cr3

	movl %cr4, %eax
	orl  $0x20, %eax
	movl %eax, %cr4

# Enable long mode in EFER.
	movl $0xC0000080, %ecx
	rdmsr
	orl  $0x100, %eax
	wrmsr

# Load GDT and enable paging.
	lgdt gdt64_desc
	movl %cr0, %eax
	orl  $0x80000000, %eax
	movl %eax, %cr0

# Far jump to 64-bit code segment.
	ljmp $0x08, $long_mode_entry

.code64
long_mode_entry:
	movw $0x10, %ax
	movw %ax, %ds
	movw %ax, %es
	movw %ax, %ss
	movw %ax, %fs
	movw %ax, %gs

	movq $stack_top, %rsp
	andq $-16, %rsp
	subq $8, %rsp

# Clear BSS.
	movq $__bss_start, %rdi
	movq $__bss_end, %rcx
	subq %rdi, %rcx
	xor %eax, %eax
	rep stosb

	movl multiboot_info_ptr(%rip), %edi
	call go_0kernel.Main

.Lhang64:
	hlt
	jmp .Lhang64
