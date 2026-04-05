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
	mov  $1, %rbx            # fd = stdout
	lea  hello_msg(%rip), %rcx
	mov  $hello_msg_len, %rdx
	int  $0x80

	mov  $2, %rax            # SYS_EXIT
	xor  %rbx, %rbx
	int  $0x80
	hlt

.global go_0kernel.userProbeReadKernelStart
go_0kernel.userProbeReadKernelStart:
	mov  $KERNEL_TEXT_BASE, %r8
	mov  (%r8), %rax         # Must #PF in ring3 (supervisor page)
	mov  $2, %rax
	mov  $11, %rbx
	int  $0x80
	hlt

.global go_0kernel.userProbeWriteKernelStart
go_0kernel.userProbeWriteKernelStart:
	mov  $KERNEL_TEXT_BASE, %r8
	movb $0x41, (%r8)        # Must #PF in ring3 (supervisor page)
	mov  $2, %rax
	mov  $12, %rbx
	int  $0x80
	hlt

hello_msg:
	.ascii "hello from userland\n"
hello_msg_end:
	.set hello_msg_len, hello_msg_end - hello_msg
