/* kernel/scheduler/switch.s */

.code64
.section .text
/* Updated mangled name for github.com/dmarro89/go-dav-os/kernel/scheduler.CpuSwitch */
.global github_0com_1dmarro89_1go_x2ddav_x2dos_1kernel_1scheduler.CpuSwitch
.type   github_0com_1dmarro89_1go_x2ddav_x2dos_1kernel_1scheduler.CpuSwitch, @function

github_0com_1dmarro89_1go_x2ddav_x2dos_1kernel_1scheduler.CpuSwitch:
	pushq %rbp
	pushq %rbx
	pushq %rsi
	pushq %rdi

	# Args: old *uint64 in RDI, new uint64 in RSI.
	movq %rsp, (%rdi)
	movq %rsi, %rsp

	popq %rdi
	popq %rsi
	popq %rbx
	popq %rbp

	ret
