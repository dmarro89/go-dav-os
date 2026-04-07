.code64
.section .text

.set USER_VA_BASE,  0x40000000
.set USER_STACK_TOP, 0x40002000

.macro PUSH_REGS
	pushq %rax
	pushq %rcx
	pushq %rdx
	pushq %rbx
	pushq %rbp
	pushq %rsi
	pushq %rdi
	pushq %r8
	pushq %r9
	pushq %r10
	pushq %r11
	pushq %r12
	pushq %r13
	pushq %r14
	pushq %r15
.endm

.macro POP_REGS
	popq %r15
	popq %r14
	popq %r13
	popq %r12
	popq %r11
	popq %r10
	popq %r9
	popq %r8
	popq %rdi
	popq %rsi
	popq %rbp
	popq %rbx
	popq %rdx
	popq %rcx
	popq %rax
.endm

# github.com/dmarro89/go-dav-os/keyboard.inb(port uint16) byte
.global github_0com_1dmarro89_1go_x2ddav_x2dos_1keyboard.inb
.type   github_0com_1dmarro89_1go_x2ddav_x2dos_1keyboard.inb, @function
github_0com_1dmarro89_1go_x2ddav_x2dos_1keyboard.inb:
	movw %di, %dx
	xorl %eax, %eax
	inb %dx, %al
	ret
.size github_0com_1dmarro89_1go_x2ddav_x2dos_1keyboard.inb, . - github_0com_1dmarro89_1go_x2ddav_x2dos_1keyboard.inb

# github.com/dmarro89/go-dav-os/keyboard.outb(port uint16, value byte)
.global github_0com_1dmarro89_1go_x2ddav_x2dos_1keyboard.outb
.type   github_0com_1dmarro89_1go_x2ddav_x2dos_1keyboard.outb, @function
github_0com_1dmarro89_1go_x2ddav_x2dos_1keyboard.outb:
	movw %di, %dx
	movb %sil, %al
	outb %al, %dx
	ret
.size github_0com_1dmarro89_1go_x2ddav_x2dos_1keyboard.outb, . - github_0com_1dmarro89_1go_x2ddav_x2dos_1keyboard.outb

# github.com/dmarro89/go-dav-os/mem.bootstrapEnd() uint64
.global github_0com_1dmarro89_1go_x2ddav_x2dos_1mem.bootstrapEnd
.type   github_0com_1dmarro89_1go_x2ddav_x2dos_1mem.bootstrapEnd, @function
github_0com_1dmarro89_1go_x2ddav_x2dos_1mem.bootstrapEnd:
	leaq __bootstrap_end(%rip), %rax
	ret
.size github_0com_1dmarro89_1go_x2ddav_x2dos_1mem.bootstrapEnd, . - github_0com_1dmarro89_1go_x2ddav_x2dos_1mem.bootstrapEnd

# github.com/dmarro89/go-dav-os/mem.kernelEnd() uint64
.global github_0com_1dmarro89_1go_x2ddav_x2dos_1mem.kernelEnd
.type   github_0com_1dmarro89_1go_x2ddav_x2dos_1mem.kernelEnd, @function
github_0com_1dmarro89_1go_x2ddav_x2dos_1mem.kernelEnd:
	leaq __kernel_end(%rip), %rax
	ret
.size github_0com_1dmarro89_1go_x2ddav_x2dos_1mem.kernelEnd, . - github_0com_1dmarro89_1go_x2ddav_x2dos_1mem.kernelEnd

# github.com/dmarro89/go-dav-os/serial.inb(port uint16) byte
.global github_0com_1dmarro89_1go_x2ddav_x2dos_1serial.inb
.type   github_0com_1dmarro89_1go_x2ddav_x2dos_1serial.inb, @function
github_0com_1dmarro89_1go_x2ddav_x2dos_1serial.inb:
	movw %di, %dx
	xorl %eax, %eax
	inb %dx, %al
	ret
.size github_0com_1dmarro89_1go_x2ddav_x2dos_1serial.inb, . - github_0com_1dmarro89_1go_x2ddav_x2dos_1serial.inb

# github.com/dmarro89/go-dav-os/serial.outb(port uint16, value byte)
.global github_0com_1dmarro89_1go_x2ddav_x2dos_1serial.outb
.type   github_0com_1dmarro89_1go_x2ddav_x2dos_1serial.outb, @function
github_0com_1dmarro89_1go_x2ddav_x2dos_1serial.outb:
	movw %di, %dx
	movb %sil, %al
	outb %al, %dx
	ret
.size github_0com_1dmarro89_1go_x2ddav_x2dos_1serial.outb, . - github_0com_1dmarro89_1go_x2ddav_x2dos_1serial.outb

# github.com/dmarro89/go-dav-os/terminal.outb(port uint16, value byte)
.global github_0com_1dmarro89_1go_x2ddav_x2dos_1terminal.outb
.type   github_0com_1dmarro89_1go_x2ddav_x2dos_1terminal.outb, @function
github_0com_1dmarro89_1go_x2ddav_x2dos_1terminal.outb:
	movw %di, %dx
	movb %sil, %al
	outb %al, %dx
	ret
.size github_0com_1dmarro89_1go_x2ddav_x2dos_1terminal.outb, . - github_0com_1dmarro89_1go_x2ddav_x2dos_1terminal.outb

# github.com/dmarro89/go-dav-os/terminal.debugChar(c byte)
.global github_0com_1dmarro89_1go_x2ddav_x2dos_1terminal.debugChar
.type   github_0com_1dmarro89_1go_x2ddav_x2dos_1terminal.debugChar, @function
github_0com_1dmarro89_1go_x2ddav_x2dos_1terminal.debugChar:
	movb %dil, %al
	outb %al, $0xe9
	ret
.size github_0com_1dmarro89_1go_x2ddav_x2dos_1terminal.debugChar, . - github_0com_1dmarro89_1go_x2ddav_x2dos_1terminal.debugChar

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
	cli
1:
	hlt
	jmp 1b
.size runtime.goPanicIndex, . - runtime.goPanicIndex

# void runtime.goPanicSliceAlen()
.global runtime.goPanicSliceAlen
.type   runtime.goPanicSliceAlen, @function
runtime.goPanicSliceAlen:
	cli
1:
	hlt
	jmp 1b
.size runtime.goPanicSliceAlen, . - runtime.goPanicSliceAlen

# void runtime.goPanicSliceB()
.global runtime.goPanicSliceB
.type   runtime.goPanicSliceB, @function
runtime.goPanicSliceB:
	cli
1:
	hlt
	jmp 1b
.size runtime.goPanicSliceB, . - runtime.goPanicSliceB

# bool runtime.panicdivide(...)
.global runtime.panicdivide
.type   runtime.panicdivide, @function
runtime.panicdivide:
	cli
1:
	hlt
	jmp 1b
.size runtime.panicdivide, . - runtime.panicdivide

# bool runtime.memequal(...)
.global runtime.memequal
.type   runtime.memequal, @function
runtime.memequal:
	xor %eax, %eax
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
	cli
1:
	hlt
	jmp 1b
.size runtime.goPanicIndexU, . - runtime.goPanicIndexU

# bool runtime.memequal32..f(...)
.global runtime.memequal32..f
.type   runtime.memequal32..f, @function
runtime.memequal32..f:
	xor %eax, %eax
	ret
.size runtime.memequal32..f, . - runtime.memequal32..f

# bool runtime.memequal16..f(...)
.global runtime.memequal16..f
.type   runtime.memequal16..f, @function
runtime.memequal16..f:
	xor %eax, %eax
	ret
.size runtime.memequal16..f, . - runtime.memequal16..f

# bool runtime.memequal8..f(...)
.global runtime.memequal8..f
.type   runtime.memequal8..f, @function
runtime.memequal8..f:
	xor %eax, %eax
	ret
.size runtime.memequal8..f, . - runtime.memequal8..f

# bool runtime.memequal64..f(...)
.global runtime.memequal64..f
.type   runtime.memequal64..f, @function
runtime.memequal64..f:
	xor %eax, %eax
	ret
.size runtime.memequal64..f, . - runtime.memequal64..f

# bool runtime.interequal..f(p, q unsafe.Pointer) bool
# Interface equality stub — keyboard.Layout type descriptor references this.
# Always returns false. Valid only because == or != are never used on non-empty
# interface values, directly or indirectly. If this changes, then this stub
# must be replaced with a real implementation.
.global runtime.interequal..f
.type   runtime.interequal..f, @function
runtime.interequal..f:
	xor %eax, %eax
	ret
.size runtime.interequal..f, . - runtime.interequal..f

# void go_0kernel.LoadIDT(void *idtr)
.global go_0kernel.LoadIDT
.type   go_0kernel.LoadIDT, @function
go_0kernel.LoadIDT:
	lidt (%rdi)
	ret

# void go_0kernel.LoadGDT(void *gdtr)
.global go_0kernel.LoadGDT
.type   go_0kernel.LoadGDT, @function
go_0kernel.LoadGDT:
	lgdt (%rdi)
	ret
.size go_0kernel.LoadGDT, . - go_0kernel.LoadGDT

# void go_0kernel.LoadTR(uint16 sel)
.global go_0kernel.LoadTR
.type   go_0kernel.LoadTR, @function
go_0kernel.LoadTR:
	movw %di, %ax
	ltr %ax
	ret
.size go_0kernel.LoadTR, . - go_0kernel.LoadTR

# void go_0kernel.LoadDataSegments(uint16 sel)
.global go_0kernel.LoadDataSegments
.type   go_0kernel.LoadDataSegments, @function
go_0kernel.LoadDataSegments:
	movw %di, %ax
	movw %ax, %ds
	movw %ax, %es
	movw %ax, %ss
	movw %ax, %fs
	movw %ax, %gs
	ret
.size go_0kernel.LoadDataSegments, . - go_0kernel.LoadDataSegments

# void go_0kernel.StoreIDT(void *idtr)
.global go_0kernel.StoreIDT
.type   go_0kernel.StoreIDT, @function
go_0kernel.StoreIDT:
	sidt (%rdi)
	ret

# void go_0kernel.Int80Stub()
.global go_0kernel.Int80Stub
.type   go_0kernel.Int80Stub, @function
go_0kernel.Int80Stub:
	PUSH_REGS
	mov %rsp, %rbp
	andq $-16, %rsp
	subq $8, %rsp
	mov %rbp, %rdi
	call  go_0kernel.Int80Handler
	mov %rbp, %rsp
	POP_REGS
	iretq
.size go_0kernel.Int80Stub, . - go_0kernel.Int80Stub

# uint64 go_0kernel.getInt80StubAddr()
.global go_0kernel.getInt80StubAddr
.type   go_0kernel.getInt80StubAddr, @function
go_0kernel.getInt80StubAddr:
	leaq go_0kernel.Int80Stub(%rip), %rax
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
	mov $0xb8000, %rdi
	movb $'G', (%rdi)
	movb $0x1f, 1(%rdi)
1:
	hlt
	jmp 1b
.size go_0kernel.GPFaultStub, . - go_0kernel.GPFaultStub

# void go_0kernel.DFaultStub()
.global go_0kernel.DFaultStub
.type   go_0kernel.DFaultStub, @function
go_0kernel.DFaultStub:
	movb $'D', %al
	cli
	mov $0xb8000, %rdi
	movb $'D', (%rdi)
	movb $0x4f, 1(%rdi)
1:
	hlt
	jmp 1b
.size go_0kernel.DFaultStub, . - go_0kernel.DFaultStub

# void go_0kernel.PFaultStub()
.global go_0kernel.PFaultStub
.type   go_0kernel.PFaultStub, @function
go_0kernel.PFaultStub:
	# Debug marker + faulting linear address from CR2.
	movb $'P', %al
	outb %al, $0xe9
	movb $'F', %al
	outb %al, $0xe9
	movb $' ', %al
	outb %al, $0xe9
	movb $'C', %al
	outb %al, $0xe9
	movb $'R', %al
	outb %al, $0xe9
	movb $'2', %al
	outb %al, $0xe9
	movb $'=', %al
	outb %al, $0xe9
	movb $'0', %al
	outb %al, $0xe9
	movb $'x', %al
	outb %al, $0xe9

	movq %cr2, %rbx
	movl $16, %ecx

.Lpf_hex_loop:
	movq %rbx, %rdx
	shrq $60, %rdx
	andb $0x0F, %dl
	cmpb $10, %dl
	jb .Lpf_hex_digit
	addb $('a' - 10), %dl
	jmp .Lpf_hex_emit

.Lpf_hex_digit:
	addb $'0', %dl

.Lpf_hex_emit:
	movb %dl, %al
	outb %al, $0xe9
	shlq $4, %rbx
	decl %ecx
	jnz .Lpf_hex_loop

	movb $'\n', %al
	outb %al, $0xe9
	cli
1:
	hlt
	jmp 1b
.size go_0kernel.PFaultStub, . - go_0kernel.PFaultStub

# uint64 go_0kernel.getGPFaultStubAddr()
.global go_0kernel.getGPFaultStubAddr
.type   go_0kernel.getGPFaultStubAddr, @function
go_0kernel.getGPFaultStubAddr:
	leaq go_0kernel.GPFaultStub(%rip), %rax
	ret
.size go_0kernel.getGPFaultStubAddr, . - go_0kernel.getGPFaultStubAddr

# uint64 go_0kernel.getDFaultStubAddr()
.global go_0kernel.getDFaultStubAddr
.type   go_0kernel.getDFaultStubAddr, @function
go_0kernel.getDFaultStubAddr:
	leaq go_0kernel.DFaultStub(%rip), %rax
	ret
.size go_0kernel.getDFaultStubAddr, . - go_0kernel.getDFaultStubAddr

# uint64 go_0kernel.getPFaultStubAddr()
.global go_0kernel.getPFaultStubAddr
.type   go_0kernel.getPFaultStubAddr, @function
go_0kernel.getPFaultStubAddr:
	leaq go_0kernel.PFaultStub(%rip), %rax
	ret
.size go_0kernel.getPFaultStubAddr, . - go_0kernel.getPFaultStubAddr

# void go_0kernel.DebugChar(byte)
.global go_0kernel.DebugChar
.type   go_0kernel.DebugChar, @function
go_0kernel.DebugChar:
	movb %dil, %al
	outb %al, $0xe9
	ret

# uint8 go_0kernel.inb(uint16 port)
.global go_0kernel.inb
.type   go_0kernel.inb, @function
go_0kernel.inb:
	movw %di, %dx
	xorl %eax, %eax
	inb %dx, %al
	ret
.size go_0kernel.inb, . - go_0kernel.inb

# void go_0kernel.outb(uint16 port, uint8 val)
.global go_0kernel.outb
.type   go_0kernel.outb, @function
go_0kernel.outb:
	movw %di, %dx
	movb %sil, %al
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
	PUSH_REGS
	mov %rsp, %rbp
	andq $-16, %rsp
	subq $8, %rsp
	call go_0kernel.IRQ0Handler
	mov %rbp, %rsp
	POP_REGS
	iretq
.size go_0kernel.IRQ0Stub, . - go_0kernel.IRQ0Stub

.global go_0kernel.getIRQ0StubAddr
.type   go_0kernel.getIRQ0StubAddr, @function
go_0kernel.getIRQ0StubAddr:
	leaq go_0kernel.IRQ0Stub(%rip), %rax
	ret
.size go_0kernel.getIRQ0StubAddr, . - go_0kernel.getIRQ0StubAddr

.global go_0kernel.IRQ1Stub
.type   go_0kernel.IRQ1Stub, @function
go_0kernel.IRQ1Stub:
	PUSH_REGS
	mov %rsp, %rbp
	andq $-16, %rsp
	subq $8, %rsp
	call go_0kernel.IRQ1Handler
	mov %rbp, %rsp
	POP_REGS
	iretq
.size go_0kernel.IRQ1Stub, . - go_0kernel.IRQ1Stub

.global go_0kernel.getIRQ1StubAddr
.type   go_0kernel.getIRQ1StubAddr, @function
go_0kernel.getIRQ1StubAddr:
	leaq go_0kernel.IRQ1Stub(%rip), %rax
	ret
.size go_0kernel.getIRQ1StubAddr, . - go_0kernel.getIRQ1StubAddr

# --- Data section: global variable runtime.writeBarrier (bool) ---
.section .data
.global  runtime.writeBarrier
.type    runtime.writeBarrier, @object
runtime.writeBarrier:
	.long 0
	.size runtime.writeBarrier, . - runtime.writeBarrier

# github.com/dmarro89/go-dav-os/drivers/ata.inb(port uint16) byte
.global github_0com_1dmarro89_1go_x2ddav_x2dos_1drivers_1ata.inb
.type   github_0com_1dmarro89_1go_x2ddav_x2dos_1drivers_1ata.inb, @function
github_0com_1dmarro89_1go_x2ddav_x2dos_1drivers_1ata.inb:
	movw %di, %dx
	xorl %eax, %eax
	inb %dx, %al
	ret
.size github_0com_1dmarro89_1go_x2ddav_x2dos_1drivers_1ata.inb, . - github_0com_1dmarro89_1go_x2ddav_x2dos_1drivers_1ata.inb

# github.com/dmarro89/go-dav-os/drivers/ata.outb(port uint16, val byte)
.global github_0com_1dmarro89_1go_x2ddav_x2dos_1drivers_1ata.outb
.type   github_0com_1dmarro89_1go_x2ddav_x2dos_1drivers_1ata.outb, @function
github_0com_1dmarro89_1go_x2ddav_x2dos_1drivers_1ata.outb:
	movw %di, %dx
	movb %sil, %al
	outb %al, %dx
	ret
.size github_0com_1dmarro89_1go_x2ddav_x2dos_1drivers_1ata.outb, . - github_0com_1dmarro89_1go_x2ddav_x2dos_1drivers_1ata.outb

# github.com/dmarro89/go-dav-os/drivers/ata.insw(port uint16, addr *byte, count int)
.global github_0com_1dmarro89_1go_x2ddav_x2dos_1drivers_1ata.insw
.type   github_0com_1dmarro89_1go_x2ddav_x2dos_1drivers_1ata.insw, @function
github_0com_1dmarro89_1go_x2ddav_x2dos_1drivers_1ata.insw:
	movq %rdx, %rcx    # count to RCX
	movw %di, %dx      # port in DX
	movq %rsi, %rdi    # addr to RDI (destination)
	cld
	rep insw
	ret
.size github_0com_1dmarro89_1go_x2ddav_x2dos_1drivers_1ata.insw, . - github_0com_1dmarro89_1go_x2ddav_x2dos_1drivers_1ata.insw

# github.com/dmarro89/go-dav-os/drivers/ata.outsw(port uint16, addr *byte, count int)
.global github_0com_1dmarro89_1go_x2ddav_x2dos_1drivers_1ata.outsw
.type   github_0com_1dmarro89_1go_x2ddav_x2dos_1drivers_1ata.outsw, @function
github_0com_1dmarro89_1go_x2ddav_x2dos_1drivers_1ata.outsw:
	movq %rdx, %rcx    # count to RCX
	movw %di, %dx      # port in DX
	# addr is already in RSI (source)
	cld
	rep outsw
	ret
.size github_0com_1dmarro89_1go_x2ddav_x2dos_1drivers_1ata.outsw, . - github_0com_1dmarro89_1go_x2ddav_x2dos_1drivers_1ata.outsw

# void go_0kernel.ExecuteUserTask(funcPtr uint64, stackPtr uint64)
.global go_0kernel.ExecuteUserTask
.type   go_0kernel.ExecuteUserTask, @function
go_0kernel.ExecuteUserTask:
    # Save kernel state
    mov %rsp, __kernel_saved_rsp(%rip)
    mov %rbp, __kernel_saved_rbp(%rip)
    mov %rbx, __kernel_saved_rbx(%rip)
    mov %r12, __kernel_saved_r12(%rip)
    mov %r13, __kernel_saved_r13(%rip)
    mov %r14, __kernel_saved_r14(%rip)
    mov %r15, __kernel_saved_r15(%rip)
    pushfq
    popq __kernel_saved_rflags(%rip)

    # Setup iretq frame
    mov $0x23, %ax      # user data selector Index 4 (0x20) | 3 = 0x23
    mov %ax, %ds
    mov %ax, %es
    mov %ax, %fs
    mov %ax, %gs

    pushq $0x23         # SS (Data)
    pushq %rsi          # RSP
    pushf               # RFLAGS
    popq %rax
    orq $0x200, %rax    # IF bit
    pushq %rax
    pushq $0x1B         # CS (Code) Index 3 (0x18) | 3 = 0x1B
    pushq %rdi          # RIP
    iretq
.size go_0kernel.ExecuteUserTask, . - go_0kernel.ExecuteUserTask

# void go_0kernel.ReturnToKernel()
.global go_0kernel.ReturnToKernel
.type   go_0kernel.ReturnToKernel, @function
go_0kernel.ReturnToKernel:
    # Restore kernel data segments
    mov $0x10, %ax
    mov %ax, %ds
    mov %ax, %es
    mov %ax, %fs
    mov %ax, %gs

    mov __kernel_saved_rsp(%rip), %rsp
    mov __kernel_saved_rbp(%rip), %rbp
    mov __kernel_saved_rbx(%rip), %rbx
    mov __kernel_saved_r12(%rip), %r12
    mov __kernel_saved_r13(%rip), %r13
    mov __kernel_saved_r14(%rip), %r14
    mov __kernel_saved_r15(%rip), %r15
    pushq __kernel_saved_rflags(%rip)
    popfq

    # Ret where ExecuteUserTask was called
    ret
.size go_0kernel.ReturnToKernel, . - go_0kernel.ReturnToKernel

# uint64 go_0kernel.GetUserProgramHelloAddr()
.global go_0kernel.GetUserProgramHelloAddr
.type   go_0kernel.GetUserProgramHelloAddr, @function
go_0kernel.GetUserProgramHelloAddr:
	leaq go_0kernel.userHelloStart(%rip), %rax
	leaq __user_program_page(%rip), %rdx
	subq %rdx, %rax
	addq $USER_VA_BASE, %rax
	ret
.size go_0kernel.GetUserProgramHelloAddr, . - go_0kernel.GetUserProgramHelloAddr

# uint64 go_0kernel.GetUserProgramKernelReadProbeAddr()
.global go_0kernel.GetUserProgramKernelReadProbeAddr
.type   go_0kernel.GetUserProgramKernelReadProbeAddr, @function
go_0kernel.GetUserProgramKernelReadProbeAddr:
	leaq go_0kernel.userProbeReadKernelStart(%rip), %rax
	leaq __user_program_page(%rip), %rdx
	subq %rdx, %rax
	addq $USER_VA_BASE, %rax
	ret
.size go_0kernel.GetUserProgramKernelReadProbeAddr, . - go_0kernel.GetUserProgramKernelReadProbeAddr

# uint64 go_0kernel.GetUserProgramKernelWriteProbeAddr()
.global go_0kernel.GetUserProgramKernelWriteProbeAddr
.type   go_0kernel.GetUserProgramKernelWriteProbeAddr, @function
go_0kernel.GetUserProgramKernelWriteProbeAddr:
	leaq go_0kernel.userProbeWriteKernelStart(%rip), %rax
	leaq __user_program_page(%rip), %rdx
	subq %rdx, %rax
	addq $USER_VA_BASE, %rax
	ret
.size go_0kernel.GetUserProgramKernelWriteProbeAddr, . - go_0kernel.GetUserProgramKernelWriteProbeAddr

# uint64 go_0kernel.GetUserStackTopAddr()
.global go_0kernel.GetUserStackTopAddr
.type   go_0kernel.GetUserStackTopAddr, @function
go_0kernel.GetUserStackTopAddr:
	movabs $USER_STACK_TOP, %rax
	ret
.size go_0kernel.GetUserStackTopAddr, . - go_0kernel.GetUserStackTopAddr


.section .data
__kernel_saved_rsp: .quad 0
__kernel_saved_rbp: .quad 0
__kernel_saved_rbx: .quad 0
__kernel_saved_r12: .quad 0
__kernel_saved_r13: .quad 0
__kernel_saved_r14: .quad 0
__kernel_saved_r15: .quad 0
__kernel_saved_rflags: .quad 0
