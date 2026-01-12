[1mdiff --git i/Makefile w/Makefile[m
[1mindex 00d9594..407442b 100644[m
[1m--- i/Makefile[m
[1m+++ w/Makefile[m
[36m@@ -24,6 +24,7 @@[m [mLINKER_SCRIPT := boot/linker.ld[m
 [m
 MODPATH          := github.com/dmarro89/go-dav-os[m
 TERMINAL_IMPORT  := $(MODPATH)/terminal[m
[32m+[m[32mSERIAL_IMPORT  := $(MODPATH)/serial[m
 KEYBOARD_IMPORT  := $(MODPATH)/keyboard[m
 SHELL_IMPORT     := $(MODPATH)/shell[m
 MEM_IMPORT     := $(MODPATH)/mem[m
[36m@@ -32,6 +33,7 @@[m [mSCHEDULER_IMPORT := $(MODPATH)/kernel/scheduler[m
 [m
 KERNEL_SRCS := $(wildcard kernel/*.go)[m
 TERMINAL_SRC := terminal/terminal.go[m
[32m+[m[32mSERIAL_SRCS := $(wildcard serial/*.go)[m
 KEYBOARD_SRCS := $(wildcard keyboard/*.go)[m
 SHELL_SRCS := $(wildcard shell/*.go)[m
 MEM_SRCS       := $(wildcard mem/*.go)[m
[36m@@ -43,6 +45,8 @@[m [mBOOT_OBJ   := $(BUILD_DIR)/boot.o[m
 KERNEL_OBJ := $(BUILD_DIR)/kernel.o[m
 TERMINAL_OBJ := $(BUILD_DIR)/terminal.o[m
 TERMINAL_GOX := $(BUILD_DIR)/github.com/dmarro89/go-dav-os/terminal.gox[m
[32m+[m[32mSERIAL_OBJ := $(BUILD_DIR)/serial.o[m
[32m+[m[32mSERIAL_GOX := $(BUILD_DIR)/github.com/dmarro89/go-dav-os/serial.gox[m
 KEYBOARD_OBJ   := $(BUILD_DIR)/keyboard.o[m
 KEYBOARD_GOX   := $(BUILD_DIR)/github.com/dmarro89/go-dav-os/keyboard.gox[m
 SHELL_OBJ   := $(BUILD_DIR)/shell.o[m
[36m@@ -82,8 +86,9 @@[m [m$(BOOT_OBJ): $(BOOT_SRCS) | $(BUILD_DIR)[m
 	$(AS) $(BOOT_SRCS) -o $(BOOT_OBJ)[m
 [m
 # --- 2. Compile terminal.go (package terminal) with gccgo ---[m
[31m-$(TERMINAL_OBJ): $(TERMINAL_SRC) | $(BUILD_DIR)[m
[32m+[m[32m$(TERMINAL_OBJ): $(TERMINAL_SRC) $(SERIAL_GOX) | $(BUILD_DIR)[m
 	$(GCCGO) -static -Werror -nostdlib -nostartfiles -nodefaultlibs \[m
[32m+[m		[32m-I $(BUILD_DIR) \[m
 		-fgo-pkgpath=$(TERMINAL_IMPORT) \[m
 		-c $(TERMINAL_SRC) -o $(TERMINAL_OBJ)[m
 [m
[36m@@ -92,6 +97,15 @@[m [m$(TERMINAL_GOX): $(TERMINAL_OBJ) | $(BUILD_DIR)[m
 	mkdir -p $(dir $(TERMINAL_GOX))[m
 	$(OBJCOPY) -j .go_export $(TERMINAL_OBJ) $(TERMINAL_GOX)[m
 [m
[32m+[m[32m# --- Serial package ---[m
[32m+[m[32m$(SERIAL_OBJ): $(SERIAL_SRCS) | $(BUILD_DIR)[m
[32m+[m	[32m$(GCCGO) -static -Werror -nostdlib -nostartfiles -nodefaultlibs \[m
[32m+[m		[32m-fgo-pkgpath=$(SERIAL_IMPORT) \[m
[32m+[m		[32m-c $(SERIAL_SRCS) -o $(SERIAL_OBJ)[m
[1;36m+[m
[32m+[m[32m$(SERIAL_GOX): $(SERIAL_OBJ) | $(BUILD_DIR)[m
[32m+[m	[32mmkdir -p $(dir $(SERIAL_GOX))[m
[32m+[m	[32m$(OBJCOPY) -j .go_export $(SERIAL_OBJ) $(SERIAL_GOX)[m
 # --- 4. Compile keyboard.go and layout.go (package keyboard) with gccgo ---[m
 $(KEYBOARD_OBJ): $(KEYBOARD_SRCS) | $(BUILD_DIR)[m
 	$(GCCGO) -static -Werror -nostdlib -nostartfiles -nodefaultlibs \[m
[36m@@ -148,7 +162,7 @@[m [m$(SCH_SWITCH_OBJ): $(SCH_SWITCH_SRC) | $(BUILD_DIR)[m
 	$(AS) $(SCH_SWITCH_SRC) -o $(SCH_SWITCH_OBJ)[m
 [m
 # --- 8. Compile kernel.go (package kernel, imports "github.com/dmarro89/go-dav-os/terminal") ---[m
[31m-$(KERNEL_OBJ): $(KERNEL_SRCS) $(TERMINAL_GOX) $(KEYBOARD_GOX) $(SHELL_GOX) $(MEM_GOX) $(FS_GOX) $(SCHEDULER_GOX) | $(BUILD_DIR)[m
[32m+[m[32m$(KERNEL_OBJ): $(KERNEL_SRCS) $(TERMINAL_GOX) $(SERIAL_GOX) $(KEYBOARD_GOX) $(SHELL_GOX) $(MEM_GOX) $(FS_GOX) $(SCHEDULER_GOX) | $(BUILD_DIR)[m
 	$(GCCGO) -static -Werror -nostdlib -nostartfiles -nodefaultlibs \[m
 		-I $(BUILD_DIR) \[m
 		-c $(KERNEL_SRCS) -o $(KERNEL_OBJ)[m
[36m@@ -156,10 +170,10 @@[m [m$(KERNEL_OBJ): $(KERNEL_SRCS) $(TERMINAL_GOX) $(KEYBOARD_GOX) $(SHELL_GOX) $(MEM[m
 # -----------------------[m
 # Link: boot.o + kernel.o -> kernel.elf[m
 # -----------------------[m
[31m-$(KERNEL_ELF): $(BOOT_OBJ) $(TERMINAL_OBJ) $(KEYBOARD_OBJ) $(SHELL_OBJ) $(MEM_OBJ) $(FS_OBJ) $(SCHEDULER_OBJ) $(SCH_SWITCH_OBJ) $(KERNEL_OBJ) $(LINKER_SCRIPT)[m
[32m+[m[32m$(KERNEL_ELF): $(BOOT_OBJ) $(TERMINAL_OBJ) $(SERIAL_OBJ) $(KEYBOARD_OBJ) $(SHELL_OBJ) $(MEM_OBJ) $(FS_OBJ) $(SCHEDULER_OBJ) $(SCH_SWITCH_OBJ) $(KERNEL_OBJ) $(LINKER_SCRIPT)[m
 	$(GCC) -T $(LINKER_SCRIPT) -o $(KERNEL_ELF) \[m
 		-ffreestanding -O2 -nostdlib \[m
[31m-		$(BOOT_OBJ) $(TERMINAL_OBJ) $(KEYBOARD_OBJ) $(SHELL_OBJ) $(MEM_OBJ) $(FS_OBJ) $(SCHEDULER_OBJ) $(SCH_SWITCH_OBJ) $(KERNEL_OBJ) -lgcc[m
[32m+[m		[32m$(BOOT_OBJ) $(TERMINAL_OBJ) $(SERIAL_OBJ) $(KEYBOARD_OBJ) $(SHELL_OBJ) $(MEM_OBJ) $(FS_OBJ) $(SCHEDULER_OBJ) $(SCH_SWITCH_OBJ) $(KERNEL_OBJ) -lgcc[m
 [m
 # -----------------------[m
 # ISO with GRUB[m
[36m@@ -192,4 +206,4 @@[m [mdocker-build-only: docker-image[m
 [m
 docker-shell: docker-image[m
 	docker run -it --rm --platform=$(DOCKER_PLATFORM) \[m
[1;35m-	  -v "$(CURDIR)":/work -w /work $(DOCKER_IMAGE) bash[m
\ No newline at end of file[m
[1;36m+[m	[1;36m  -v "$(CURDIR)":/work -w /work $(DOCKER_IMAGE) bash[m
[1mdiff --git i/boot/boot.s w/boot/boot.s[m
[1mindex 03ece21..cb88821 100644[m
[1m--- i/boot/boot.s[m
[1m+++ w/boot/boot.s[m
[36m@@ -105,6 +105,35 @@[m [mgithub_0com_1dmarro89_1go_x2ddav_x2dos_1keyboard.outb:[m
 	ret[m
 .size github_0com_1dmarro89_1go_x2ddav_x2dos_1keyboard.outb, . - github_0com_1dmarro89_1go_x2ddav_x2dos_1keyboard.outb[m
 [m
[32m+[m[32m# --------------------------------------------------[m
[32m+[m[32m# github_0com_1dmarro89_1go_x2ddav_x2dos_1serial.inb(port uint16) byte[m
[32m+[m[32m# arg0 (port) at 4(%esp), return in %al[m
[32m+[m[32m# --------------------------------------------------[m
[32m+[m[32m.global github_0com_1dmarro89_1go_x2ddav_x2dos_1serial.inb[m
[32m+[m[32m.type   github_0com_1dmarro89_1go_x2ddav_x2dos_1serial.inb, @function[m
[1;36m+[m
[32m+[m[32mgithub_0com_1dmarro89_1go_x2ddav_x2dos_1serial.inb:[m
[32m+[m	[32mmov 4(%esp), %dx       # port[m
[32m+[m	[32mxor %eax, %eax[m
[32m+[m	[32minb %dx, %al           # read byte from port into AL[m
[32m+[m	[32mret[m
[32m+[m[32m.size github_0com_1dmarro89_1go_x2ddav_x2dos_1serial.inb, . - github_0com_1dmarro89_1go_x2ddav_x2dos_1serial.inb[m
[1;36m+[m
[32m+[m[32m# --------------------------------------------------[m
[32m+[m[32m# github_0com_1dmarro89_1go_x2ddav_x2dos_1serial.outb(port uint16, value byte)[m
[32m+[m[32m# arg0: port  at 4(%esp)[m
[32m+[m[32m# arg1: value at 8(%esp)[m
[32m+[m[32m# --------------------------------------------------[m
[32m+[m[32m.global github_0com_1dmarro89_1go_x2ddav_x2dos_1serial.outb[m
[32m+[m[32m.type   github_0com_1dmarro89_1go_x2ddav_x2dos_1serial.outb, @function[m
[1;36m+[m
[32m+[m[32mgithub_0com_1dmarro89_1go_x2ddav_x2dos_1serial.outb:[m
[32m+[m	[32mmov  4(%esp), %dx       # port[m
[32m+[m	[32mmov  8(%esp), %al       # value[m
[32m+[m	[32moutb %al, %dx[m
[32m+[m	[32mret[m
[32m+[m[32m.size github_0com_1dmarro89_1go_x2ddav_x2dos_1serial.outb, . - github_0com_1dmarro89_1go_x2ddav_x2dos_1serial.outb[m
[1;36m+[m
 # __go_register_gc_roots(void)[m
 .global __go_register_gc_roots[m
 .type   __go_register_gc_roots, @function[m
[36m@@ -463,4 +492,3 @@[m [mgo_0kernel.getIRQ1StubAddr:[m
 runtime.writeBarrier:[m
 	.long 0    # false: GC write barrier disabled[m
 	.size runtime.writeBarrier, . - runtime.writeBarrier[m
[1;35m-[m
[1mdiff --git i/kernel/kernel.go w/kernel/kernel.go[m
[1mindex bcd796d..3ac1000 100644[m
[1m--- i/kernel/kernel.go[m
[1m+++ w/kernel/kernel.go[m
[36m@@ -5,6 +5,7 @@[m [mimport ([m
 	"github.com/dmarro89/go-dav-os/kernel/scheduler"[m
 	"github.com/dmarro89/go-dav-os/keyboard"[m
 	"github.com/dmarro89/go-dav-os/mem"[m
[32m+[m	[32m"github.com/dmarro89/go-dav-os/serial"[m
 	"github.com/dmarro89/go-dav-os/shell"[m
 	"github.com/dmarro89/go-dav-os/terminal"[m
 )[m
[36m@@ -18,6 +19,7 @@[m [mfunc Halt()[m
 [m
 func Main(multibootInfoAddr uint32) {[m
 	DisableInterrupts()[m
[32m+[m	[32mserial.Init()[m
 	terminal.Init()[m
 	terminal.Clear()[m
 [m
[36m@@ -34,8 +36,8 @@[m [mfunc Main(multibootInfoAddr uint32) {[m
 	mem.InitPFA()[m
 [m
 	scheduler.Init()[m
[31m-	scheduler.NewTask(taskA)[m
[31m-	scheduler.NewTask(taskB)[m
[32m+[m	[32m//scheduler.NewTask(taskA)[m
[32m+[m	[32m//scheduler.NewTask(taskB)[m
 [m
 	fs.Init()[m
 [m
[36m@@ -47,6 +49,10 @@[m [mfunc Main(multibootInfoAddr uint32) {[m
 		r, ok := keyboard.TryRead()[m
 		EnableInterrupts()[m
 		if !ok {[m
[32m+[m			[32mif sr, sok := serial.TryRead(); sok {[m
[32m+[m				[32mshell.FeedRune(sr)[m
[32m+[m				[32mcontinue[m
[32m+[m			[32m}[m
 			Halt()[m
 			continue[m
 		}[m
[1mdiff --git i/terminal/terminal.go w/terminal/terminal.go[m
[1mindex ae3e13c..a371599 100644[m
[1m--- i/terminal/terminal.go[m
[1m+++ w/terminal/terminal.go[m
[36m@@ -1,6 +1,10 @@[m
 package terminal[m
 [m
[31m-import "unsafe"[m
[32m+[m[32mimport ([m
[32m+[m	[32m"unsafe"[m
[1;36m+[m
[32m+[m	[32m"github.com/dmarro89/go-dav-os/serial"[m
[32m+[m[32m)[m
 [m
 func outb(port uint16, value byte)[m
 func debugChar(c byte)[m
[36m@@ -66,6 +70,10 @@[m [mfunc putRune(ch rune) {[m
 	}[m
 [m
 	if ch == '\n' {[m
[32m+[m		[32mif serial.Enabled() {[m
[32m+[m			[32mserial.WriteByte('\r')[m
[32m+[m			[32mserial.WriteByte('\n')[m
[32m+[m		[32m}[m
 		column = 0[m
 		row++[m
 		if row >= VGAHeight {[m
[36m@@ -80,6 +88,9 @@[m [mfunc putRune(ch rune) {[m
 	vidMem[row][column][1] = color[m
 [m
 	debugChar(byte(ch))[m
[32m+[m	[32mif serial.Enabled() {[m
[32m+[m		[32mserial.WriteByte(byte(ch))[m
[32m+[m	[32m}[m
 [m
 	column++[m
 	if column >= VGAWidth {[m
[36m@@ -156,6 +167,11 @@[m [mfunc putRuneAt(col, currRow int, ch rune) {[m
 }[m
 [m
 func Backspace() {[m
[32m+[m	[32mif serial.Enabled() {[m
[32m+[m		[32mserial.WriteByte('\b')[m
[32m+[m		[32mserial.WriteByte(' ')[m
[32m+[m		[32mserial.WriteByte('\b')[m
[32m+[m	[32m}[m
 	if column > 0 {[m
 		column--[m
 		vidMem[row][column][0] = ' '[m
