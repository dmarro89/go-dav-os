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
pd0:
	.skip 4096
.align 4096
pd1:
	.skip 4096
.align 4096
pd2:
	.skip 4096
.align 4096
pd3:
	.skip 4096
.global __bootstrap_end
__bootstrap_end:

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
# Build minimal identity-mapped paging (4 GiB via 2 MiB pages).
	lea pml4, %edi
	movl $pdpt, %eax
	orl $0x03, %eax
	movl %eax, (%edi)
	movl $0, 4(%edi)

	lea pdpt, %edi
	movl $pd0, %eax
	orl $0x03, %eax
	movl %eax, (%edi)
	movl $0, 4(%edi)
	movl $pd1, %eax
	orl $0x03, %eax
	movl %eax, 8(%edi)
	movl $0, 12(%edi)
	movl $pd2, %eax
	orl $0x03, %eax
	movl %eax, 16(%edi)
	movl $0, 20(%edi)
	movl $pd3, %eax
	orl $0x03, %eax
	movl %eax, 24(%edi)
	movl $0, 28(%edi)

	movl $0x83, %edx           # present|rw|ps

	lea pd0, %edi
	xorl %ecx, %ecx
	movl $0x00000000, %ebx

.Lmap_2m_pd0:
	movl %ecx, %eax
	shll $21, %eax             # ecx * 2 MiB
	addl %ebx, %eax
	orl %edx, %eax
	movl %eax, (%edi)
	movl $0, 4(%edi)
	addl $8, %edi
	incl %ecx
	cmpl $512, %ecx
	jne .Lmap_2m_pd0

	lea pd1, %edi
	xorl %ecx, %ecx
	movl $0x40000000, %ebx

.Lmap_2m_pd1:
	movl %ecx, %eax
	shll $21, %eax
	addl %ebx, %eax
	orl %edx, %eax
	movl %eax, (%edi)
	movl $0, 4(%edi)
	addl $8, %edi
	incl %ecx
	cmpl $512, %ecx
	jne .Lmap_2m_pd1

	lea pd2, %edi
	xorl %ecx, %ecx
	movl $0x80000000, %ebx

.Lmap_2m_pd2:
	movl %ecx, %eax
	shll $21, %eax
	addl %ebx, %eax
	orl %edx, %eax
	movl %eax, (%edi)
	movl $0, 4(%edi)
	addl $8, %edi
	incl %ecx
	cmpl $512, %ecx
	jne .Lmap_2m_pd2

	lea pd3, %edi
	xorl %ecx, %ecx
	movl $0xC0000000, %ebx

.Lmap_2m_pd3:
	movl %ecx, %eax
	shll $21, %eax
	addl %ebx, %eax
	orl %edx, %eax
	movl %eax, (%edi)
	movl $0, 4(%edi)
	addl $8, %edi
	incl %ecx
	cmpl $512, %ecx
	jne .Lmap_2m_pd3

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

# void runtime.goPanicSliceAlen()
.global runtime.goPanicSliceAlen
.type   runtime.goPanicSliceAlen, @function

runtime.goPanicSliceAlen:
	cli

4:
	hlt
	jmp 4b
.size runtime.goPanicSliceAlen, . - runtime.goPanicSliceAlen

# void runtime.goPanicSliceB()
.global runtime.goPanicSliceB
.type   runtime.goPanicSliceB, @function

runtime.goPanicSliceB:
	cli

5:
	hlt
	jmp 5b
.size runtime.goPanicSliceB, . - runtime.goPanicSliceB

# bool runtime.panicdivide(...)
# For now we just return 0 (not equal) to satisfy the linker.
.global runtime.panicdivide
.type   runtime.panicdivide, @function

runtime.panicdivide:
	cli

6:
	hlt
	jmp 5b
.size runtime.panicdivide, . - runtime.panicdivide

# bool runtime.memequal(...)
# For now we just return 0 (not equal) to satisfy the linker.
.global runtime.memequal
.type   runtime.memequal, @function

runtime.memequal:
	xor %eax, %eax   # return 0
	ret
.size runtime.memequal, . - runtime.memequal

.global runtime.panicmem
runtime.panicmem:
    cli
1:
    hlt
    jmp 1b
	
# void runtime.registerGCRoots()
.global runtime.registerGCRoots
.type   runtime.registerGCRoots, @function

runtime.registerGCRoots:
	ret
.size runtime.registerGCRoots, . - runtime.registerGCRoots

# void runtime.goPanicIndexU()
.global runtime.goPanicIndexU
.type   runtime.goPanicIndexU, @function

runtime.goPanicIndexU:
# If we ever hit an index-out-of-range unsigned, just halt forever for now.
cli

1:
	hlt
	jmp 1b
.size runtime.goPanicIndexU, . - runtime.goPanicIndexU

# bool runtime.memequal32..f(...)
# For now we just return 0 (not equal) to satisfy the linker.
.global runtime.memequal32..f
.type   runtime.memequal32..f, @function

runtime.memequal32..f:
	xor %eax, %eax   # return 0
	ret
.size runtime.memequal32..f, . - runtime.memequal32..f

# bool runtime.memequal16..f(...)
# For now we just return 0 (not equal) to satisfy the linker.
.global runtime.memequal16..f
.type   runtime.memequal16..f, @function

runtime.memequal16..f:
	xor %eax, %eax        # false
	ret
.size runtime.memequal16..f, . - runtime.memequal16..f

# bool runtime.memequal8..f(...)
# For now we just return 0 (not equal) to satisfy the linker.
.global runtime.memequal8..f
.type   runtime.memequal8..f, @function

runtime.memequal8..f:
	xor %eax, %eax        # false
	ret
.size runtime.memequal8..f, . - runtime.memequal8..f

# github.com/dmarro89/go-dav-os/terminal.outb(port uint16, value byte)
.global github_0com_1dmarro89_1go_x2ddav_x2dos_1terminal.outb
.type   github_0com_1dmarro89_1go_x2ddav_x2dos_1terminal.outb, @function

github_0com_1dmarro89_1go_x2ddav_x2dos_1terminal.outb:
	mov  4(%esp), %dx       # port
	mov  8(%esp), %al       # value
	outb %al, %dx
	ret
.size github_0com_1dmarro89_1go_x2ddav_x2dos_1terminal.outb, . - github_0com_1dmarro89_1go_x2ddav_x2dos_1terminal.outb

# void go_0kernel.LoadIDT(uint32 *idtr)
.global go_0kernel.LoadIDT
.type   go_0kernel.LoadIDT, @function

go_0kernel.LoadIDT:
	mov  4(%esp), %eax
	lidt (%eax)          # eax -> [6]byte packed
	ret

# void go_0kernel.StoreIDT(uint32 *idtr)
.global go_0kernel.StoreIDT
.type   go_0kernel.StoreIDT, @function

go_0kernel.StoreIDT:
	mov  4(%esp), %eax
	sidt (%eax)          # write 6 bytes
	ret

# void go_0kernel.Int80Stub()
.global go_0kernel.Int80Stub
.type   go_0kernel.Int80Stub, @function

go_0kernel.Int80Stub:
    pusha

    # push argument: pointer to TrapFrame (current ESP)
    pushl %esp
    call  go_0kernel.Int80Handler
    add   $4, %esp

    popa
    iret
.size go_0kernel.Int80Stub, . - go_0kernel.Int80Stub

# void go_0kernel.TriggerSysWrite(uint32 buf, uint32 n)
.global go_0kernel.TriggerSysWrite
.type   go_0kernel.TriggerSysWrite, @function

go_0kernel.TriggerSysWrite:
    mov  4(%esp), %ecx   # buf
    mov  8(%esp), %edx   # n
    mov  $1, %eax        # SYS_WRITE
    mov  $1, %ebx        # fd=1
    int  $0x80
    ret
.size go_0kernel.TriggerSysWrite, . - go_0kernel.TriggerSysWrite

# uint32 go_0kernel.getInt80StubAddr()
.global go_0kernel.getInt80StubAddr
.type   go_0kernel.getInt80StubAddr, @function

go_0kernel.getInt80StubAddr:
	mov $go_0kernel.Int80Stub, %eax
	ret
.size go_0kernel.getInt80StubAddr, . - go_0kernel.getInt80StubAddr

# uint16 go_0kernel.GetCS()
.global go_0kernel.GetCS
.type   go_0kernel.GetCS, @function

go_0kernel.GetCS:
	mov %cs, %ax
	ret
.size go_0kernel.GetCS, . - go_0kernel.GetCS

# void go_0kernel.TriggerInt80()
.global go_0kernel.TriggerInt80
.type   go_0kernel.TriggerInt80, @function

go_0kernel.TriggerInt80:
	int $0x80
	ret
.size go_0kernel.TriggerInt80, . - go_0kernel.TriggerInt80

# void go_0kernel.GPFaultStub()
.global go_0kernel.GPFaultStub
.type   go_0kernel.GPFaultStub, @function

go_0kernel.GPFaultStub:
	movb $'G', %al
	cli
	mov  $0xb8000, %edi
	movb $'G', (%edi)
	movb $0x1f, 1(%edi)

1:
	hlt
	jmp 1b
.size go_0kernel.GPFaultStub, . - go_0kernel.GPFaultStub

.global go_0kernel.DFaultStub
.type   go_0kernel.DFaultStub, @function

# void go_0kernel.DFaultStub()
go_0kernel.DFaultStub:
	movb $'D', %al
	cli
	mov  $0xb8000, %edi
	movb $'D', (%edi)
	movb $0x4f, 1(%edi)

1:
	hlt
	jmp 1b
.size go_0kernel.DFaultStub, . - go_0kernel.DFaultStub

# uint32 go_0kernel.getGPFaultStubAddr()
.global go_0kernel.getGPFaultStubAddr
.type   go_0kernel.getGPFaultStubAddr, @function

go_0kernel.getGPFaultStubAddr:
	mov $go_0kernel.GPFaultStub, %eax
	ret
.size go_0kernel.getGPFaultStubAddr, . - go_0kernel.getGPFaultStubAddr

# uint32 go_0kernel.getDFaultStubAddr()
.global go_0kernel.getDFaultStubAddr
.type   go_0kernel.getDFaultStubAddr, @function

go_0kernel.getDFaultStubAddr:
	mov $go_0kernel.DFaultStub, %eax
	ret
.size go_0kernel.getDFaultStubAddr, . - go_0kernel.getDFaultStubAddr

# void go_0kernel.DebugChar(byte)
.global go_0kernel.DebugChar
.type   go_0kernel.DebugChar, @function

go_0kernel.DebugChar:
	mov  4(%esp), %eax       # al = arg (byte), prendiamo dal low8
	outb %al, $0xe9
	ret

# uint8  go_0kernel.inb(uint16 port)
.global go_0kernel.inb
.type   go_0kernel.inb, @function
go_0kernel.inb:
    mov 4(%esp), %dx
    xor %eax, %eax
    inb %dx, %al
    ret
.size go_0kernel.inb, . - go_0kernel.inb

# void go_0kernel.outb(uint16 port, uint8 val)
.global go_0kernel.outb
.type   go_0kernel.outb, @function
go_0kernel.outb:
    mov 4(%esp), %dx
    mov 8(%esp), %al
    outb %al, %dx
    ret
.size go_0kernel.outb, . - go_0kernel.outb

.global go_0kernel.EnableInterrupts
.type   go_0kernel.EnableInterrupts, @function
go_0kernel.EnableInterrupts:
    sti
    ret
.size go_0kernel.EnableInterrupts, . - go_0kernel.EnableInterrupts

.global go_0kernel.DisableInterrupts
.type   go_0kernel.DisableInterrupts, @function
go_0kernel.DisableInterrupts:
    cli
    ret
.size go_0kernel.DisableInterrupts, . - go_0kernel.DisableInterrupts

.global go_0kernel.Halt
.type   go_0kernel.Halt, @function
go_0kernel.Halt:
	hlt
	ret
.size go_0kernel.Halt, . - go_0kernel.Halt

.global go_0kernel.IRQ0Stub
.type   go_0kernel.IRQ0Stub, @function
go_0kernel.IRQ0Stub:
    pusha
    call go_0kernel.IRQ0Handler
    popa
    iret
.size go_0kernel.IRQ0Stub, . - go_0kernel.IRQ0Stub

.global go_0kernel.getIRQ0StubAddr
.type   go_0kernel.getIRQ0StubAddr, @function
go_0kernel.getIRQ0StubAddr:
    mov $go_0kernel.IRQ0Stub, %eax
    ret
.size go_0kernel.getIRQ0StubAddr, . - go_0kernel.getIRQ0StubAddr

.global go_0kernel.IRQ1Stub
.type   go_0kernel.IRQ1Stub, @function
go_0kernel.IRQ1Stub:
    pusha
    call go_0kernel.IRQ1Handler
    popa
    iret
.size go_0kernel.IRQ1Stub, . - go_0kernel.IRQ1Stub

.global go_0kernel.getIRQ1StubAddr
.type   go_0kernel.getIRQ1StubAddr, @function
go_0kernel.getIRQ1StubAddr:
    mov $go_0kernel.IRQ1Stub, %eax
    ret
.size go_0kernel.getIRQ1StubAddr, . - go_0kernel.getIRQ1StubAddr

// --- Data section: global variable runtime.writeBarrier (bool) ---
.section .data
.global  runtime.writeBarrier
.type    runtime.writeBarrier, @object

runtime.writeBarrier:
	.long 0    # false: GC write barrier disabled
	.size runtime.writeBarrier, . - runtime.writeBarrier

