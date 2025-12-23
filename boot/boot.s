/* boot/boot.s
 * Kernel entry point + Multiboot header for GRUB.
 *
 * Flow (Multiboot context):
 * - Provide the Multiboot header (magic 0x1BADB002, flags=0, checksum) so GRUB
 *   recognizes and loads this image.
 * - GRUB jumps to _start with EAX=0x2BADB002 and EBX pointing to the multiboot
 *   info struct (per spec).
 * - We immediately disable interrupts (cli) because no IDT/PIC is set up yet.
 * - We set ESP to a known 16 KB stack in .bss, aligned to 16 bytes.
 * - We park the CPU in a HLT loop as a placeholder until real kernel init runs.
 */
/* Multiboot header fields: GRUB expects magic, flags, and a checksum such that
 * magic + flags + checksum == 0 (32-bit). Flags=0 means no extra requirements
 * (e.g., no memory maps requested here).
 */
.set MULTIBOOT_MAGIC, 0x1BADB002
.set MULTIBOOT_FLAGS, 0x0
.set MULTIBOOT_CHECKSUM, -(MULTIBOOT_MAGIC + MULTIBOOT_FLAGS)

/* ---------------------------
 * Multiboot header section
 * ---------------------------
 * Placed in its own section so the linker keeps these 12 bytes contiguous and
 * 4-byte aligned. GRUB scans for this header to accept the image and jump to
 * the entry point with the Multiboot registers set.
 */
.section .multiboot
.align   4
.long    MULTIBOOT_MAGIC
.long    MULTIBOOT_FLAGS
.long    MULTIBOOT_CHECKSUM

// --- Stack ---
.section .bootstrap_stack, "aw", @nobits

stack_bottom:
	.skip 16384              # 16 KB di stack

stack_top:

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
call go_0kernel.Main
cli

.Lhang:
	jmp .Lhang
.size _start, . - _start

/* -----------------------------------------------------------------
 * Minimal runtime stubs required by gccgo in freestanding mode
* ----------------------------------------------------------------- */
.section .text

# --------------------------------------------------
# github_0com_1dmarro89_1go_x2ddav_x2dos_1keyboard.inb(port uint16) byte
# arg0 (port) at 4(%esp), return in %al
# --------------------------------------------------
.global github_0com_1dmarro89_1go_x2ddav_x2dos_1keyboard.inb
.type   github_0com_1dmarro89_1go_x2ddav_x2dos_1keyboard.inb, @function

github_0com_1dmarro89_1go_x2ddav_x2dos_1keyboard.inb:
	mov 4(%esp), %dx       # port
	xor %eax, %eax
	inb %dx, %al           # read byte from port into AL
	ret
.size github_0com_1dmarro89_1go_x2ddav_x2dos_1keyboard.inb, . - github_0com_1dmarro89_1go_x2ddav_x2dos_1keyboard.inb

# --------------------------------------------------
# github_0com_1dmarro89_1go_x2ddav_x2dos_1keyboard.outb(port uint16, value byte)
# arg0: port  at 4(%esp)
# arg1: value at 8(%esp)
# --------------------------------------------------
.global github_0com_1dmarro89_1go_x2ddav_x2dos_1keyboard.outb
.type   github_0com_1dmarro89_1go_x2ddav_x2dos_1keyboard.outb, @function

github_0com_1dmarro89_1go_x2ddav_x2dos_1keyboard.outb:
	mov  4(%esp), %dx       # port
	mov  8(%esp), %al       # value
	outb %al, %dx
	ret
.size github_0com_1dmarro89_1go_x2ddav_x2dos_1keyboard.outb, . - github_0com_1dmarro89_1go_x2ddav_x2dos_1keyboard.outb

# __go_register_gc_roots(void)
.global __go_register_gc_roots
.type   __go_register_gc_roots, @function

__go_register_gc_roots:
	ret
.size __go_register_gc_roots, . - __go_register_gc_roots

# __go_runtime_error(void)
.global __go_runtime_error
.type   __go_runtime_error, @function

__go_runtime_error:
	ret
.size __go_runtime_error, . - __go_runtime_error

# void runtime.gcWriteBarrier()
.global runtime.gcWriteBarrier
.type   runtime.gcWriteBarrier, @function

runtime.gcWriteBarrier:
	ret
.size runtime.gcWriteBarrier, . - runtime.gcWriteBarrier

# void runtime.goPanicIndex()
.global runtime.goPanicIndex
.type   runtime.goPanicIndex, @function

runtime.goPanicIndex:
# If we ever hit an index-out-of-range, just halt forever for now.
cli

2:
	hlt
	jmp 2b
.size runtime.goPanicIndex, . - runtime.goPanicIndex

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
	call  go_0kernel.Int80Handler
	popa
	iret
	.size go_0kernel.Int80Stub, . - go_0kernel.Int80Stub

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

