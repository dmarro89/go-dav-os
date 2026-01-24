.code64
.section .text
.global memcmp
.type memcmp, @function
memcmp:
	test %rdx, %rdx
	je .memcmp_equal

.memcmp_loop:
	movzbl (%rdi), %eax
	movzbl (%rsi), %ecx
	cmp %ecx, %eax
	jne .memcmp_diff
	inc %rdi
	inc %rsi
	dec %rdx
	jne .memcmp_loop

.memcmp_equal:
	xorl %eax, %eax
	ret

.memcmp_diff:
	subl %ecx, %eax
	ret
