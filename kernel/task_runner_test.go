package kernel

import (
	"os"
	"strings"
	"testing"
)

func TestUserHelloUsesSyscallABI(t *testing.T) {
	// Verifies that user/hello.s uses the new syscall ABI
	content, err := os.ReadFile("../user/hello.s")
	if err != nil {
		t.Fatalf("failed to read hello.s: %v", err)
	}

	src := string(content)
	if !strings.Contains(src, "syscall") {
		t.Error("hello.s should use 'syscall' instruction for the new ABI")
	}
	if strings.Contains(src, "int $0x80") || strings.Contains(src, "int 0x80") {
		t.Error("hello.s should not use the old 'int 0x80' ABI")
	}

	// Check for correct register usage for SYS_WRITE (1)
	if !strings.Contains(src, "mov  $1, %rax") || !strings.Contains(src, "mov  $1, %rdi") {
		t.Error("hello.s should use %rax for syscall number and %rdi for fd in SYS_WRITE")
	}

	// Check for correct register usage for SYS_EXIT (2)
	if !strings.Contains(src, "mov  $2, %rax") || !strings.Contains(src, "xor  %rdi, %rdi") {
		t.Error("hello.s should use %rax for syscall number and %rdi for status in SYS_EXIT")
	}
}
