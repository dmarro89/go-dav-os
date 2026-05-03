# hello.s - ring3 user programs mapped into the dedicated user VA range.

.set KERNEL_TEXT_BASE, 0x00100000

.code64
.section .user_prog, "ax"
.align 4096
.global __user_program_page
__user_program_page:

.global go_0kernel.userHelloStart
go_0kernel.userHelloStart:
	mov  $1, %rax            # SYS_WRITE
	mov  $1, %rdi            # fd = stdout
	lea  hello_msg(%rip), %rsi
	mov  $hello_msg_len, %rdx
	syscall

	mov  $2, %rax            # SYS_EXIT
	xor  %rdi, %rdi
	syscall
	hlt

.global go_0kernel.userProbeReadKernelStart
go_0kernel.userProbeReadKernelStart:
	mov  $KERNEL_TEXT_BASE, %r8
	mov  (%r8), %rax         # Must #PF in ring3 (supervisor page)
	mov  $2, %rax
	mov  $11, %rdi
	syscall
	hlt

.global go_0kernel.userProbeWriteKernelStart
go_0kernel.userProbeWriteKernelStart:
	mov  $KERNEL_TEXT_BASE, %r8
	movb $0x41, (%r8)        # Must #PF in ring3 (supervisor page)
	mov  $2, %rax
	mov  $12, %rdi
	syscall
	hlt

hello_msg:
	.ascii "hello from userland\n"
hello_msg_end:
	.set hello_msg_len, hello_msg_end - hello_msg

.global go_0kernel.userProbePrivilegedStart
go_0kernel.userProbePrivilegedStart:
	cli              # Privileged instruction, must #GP in ring 3
	hlt
