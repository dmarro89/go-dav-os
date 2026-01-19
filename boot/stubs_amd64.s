.code64
.section .text

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

# void go_0kernel.LoadIDT(void *idtr)
.global go_0kernel.LoadIDT
.type   go_0kernel.LoadIDT, @function
go_0kernel.LoadIDT:
	lidt (%rdi)
	ret

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
