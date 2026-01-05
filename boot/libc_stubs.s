.global memcmp
memcmp:
    pushl %ebp
    movl  %esp, %ebp

    movl  8(%ebp), %esi
    movl 12(%ebp), %edi
    movl 16(%ebp), %ecx

    testl %ecx, %ecx
    je    .memcmp_equal

.memcmp_loop:
    movzbl (%esi), %eax
    movzbl (%edi), %edx
    cmpl  %edx, %eax
    jne   .memcmp_diff
    incl  %esi
    incl  %edi
    decl  %ecx
    jne   .memcmp_loop

.memcmp_equal:
    xorl  %eax, %eax
    popl  %ebp
    ret

.memcmp_diff:
    subl  %edx, %eax
    popl  %ebp
    ret