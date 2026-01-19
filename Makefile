CROSS    ?= x86_64-elf

AS       := $(CROSS)-as
GCC      := $(CROSS)-gcc
GCCGO    := $(CROSS)-gccgo
OBJCOPY  := $(CROSS)-objcopy
GCCGOFLAGS := -m64
GRUB_CFG      := iso/grub/grub.cfg

GRUBMKRESCUE  := grub-mkrescue
QEMU          := qemu-system-x86_64

DOCKER_PLATFORM := linux/amd64
DOCKER_IMAGE    := go-dav-os-toolchain
DOCKER_RUN_FLAGS=-it

BUILD_DIR := build
ISO_DIR   := $(BUILD_DIR)/isodir

KERNEL_ELF := $(BUILD_DIR)/kernel.elf
ISO_IMAGE   := $(BUILD_DIR)/dav-go-os.iso

BOOT_SRCS := $(wildcard boot/*.s)
LINKER_SCRIPT := boot/linker.ld

MODPATH          := github.com/dmarro89/go-dav-os
TERMINAL_IMPORT  := $(MODPATH)/terminal
KEYBOARD_IMPORT  := $(MODPATH)/keyboard
SHELL_IMPORT     := $(MODPATH)/shell
MEM_IMPORT     := $(MODPATH)/mem
FS_IMPORT := $(MODPATH)/fs
SCHEDULER_IMPORT := $(MODPATH)/kernel/scheduler

KERNEL_SRCS := $(filter-out %_test.go, $(wildcard kernel/*.go))
TERMINAL_SRC := terminal/terminal.go
KEYBOARD_SRCS := $(filter-out %_test.go, $(wildcard keyboard/*.go))
SHELL_SRCS := $(filter-out %_test.go, $(wildcard shell/*.go))
MEM_SRCS       := $(filter-out %_test.go, $(wildcard mem/*.go))
FS_SRCS   := $(filter-out %_test.go, $(wildcard fs/*.go))
SCHEDULER_SRCS := $(filter-out %_test.go, $(wildcard kernel/scheduler/*.go))
SCH_SWITCH_SRC := kernel/scheduler/switch.s

BOOT_OBJ   := $(BUILD_DIR)/boot.o
KERNEL_OBJ := $(BUILD_DIR)/kernel.o
TERMINAL_OBJ := $(BUILD_DIR)/terminal.o
TERMINAL_GOX := $(BUILD_DIR)/github.com/dmarro89/go-dav-os/terminal.gox
KEYBOARD_OBJ   := $(BUILD_DIR)/keyboard.o
KEYBOARD_GOX   := $(BUILD_DIR)/github.com/dmarro89/go-dav-os/keyboard.gox
SHELL_OBJ   := $(BUILD_DIR)/shell.o
SHELL_GOX   := $(BUILD_DIR)/github.com/dmarro89/go-dav-os/shell.gox
MEM_OBJ   := $(BUILD_DIR)/mem.o
MEM_GOX        := $(BUILD_DIR)/github.com/dmarro89/go-dav-os/mem.gox
FS_OBJ    := $(BUILD_DIR)/fs.o
FS_GOX    := $(BUILD_DIR)/github.com/dmarro89/go-dav-os/fs.gox
SCHEDULER_OBJ := $(BUILD_DIR)/scheduler.o
SCH_SWITCH_OBJ := $(BUILD_DIR)/switch.o
SCHEDULER_GOX := $(BUILD_DIR)/github.com/dmarro89/go-dav-os/kernel/scheduler.gox

.PHONY: all kernel iso run clean docker-build docker-shell docker-run

all: $(ISO_IMAGE)

kernel: $(KERNEL_ELF)

iso: $(ISO_IMAGE)

run: $(ISO_IMAGE)
	$(QEMU) -cdrom $(ISO_IMAGE)

clean:
	rm -rf $(BUILD_DIR)

# -----------------------
# Build directory
# -----------------------
$(BUILD_DIR):
	mkdir -p $(BUILD_DIR)

# -----------------------
# Assembly: boot.s -> boot.o
# -----------------------
$(BOOT_OBJ): $(BOOT_SRCS) | $(BUILD_DIR)
	$(AS) $(BOOT_SRCS) -o $(BOOT_OBJ)

# --- 2. Compile terminal.go (package terminal) with gccgo ---
$(TERMINAL_OBJ): $(TERMINAL_SRC) | $(BUILD_DIR)
	$(GCCGO) $(GCCGOFLAGS) -static -Werror -nostdlib -nostartfiles -nodefaultlibs \
		-fgo-pkgpath=$(TERMINAL_IMPORT) \
		-c $(TERMINAL_SRC) -o $(TERMINAL_OBJ)

# --- 3. Extract .go_export into terminal.gox ---
$(TERMINAL_GOX): $(TERMINAL_OBJ) | $(BUILD_DIR)
	mkdir -p $(dir $(TERMINAL_GOX))
	$(OBJCOPY) -j .go_export $(TERMINAL_OBJ) $(TERMINAL_GOX)

# --- 4. Compile keyboard.go and layout.go (package keyboard) with gccgo ---
$(KEYBOARD_OBJ): $(KEYBOARD_SRCS) | $(BUILD_DIR)
	$(GCCGO) $(GCCGOFLAGS) -static -Werror -nostdlib -nostartfiles -nodefaultlibs \
		-fgo-pkgpath=$(KEYBOARD_IMPORT) \
		-c $(KEYBOARD_SRCS) -o $(KEYBOARD_OBJ)

# --- 5. Extract .go_export into keyboard.gox ---
$(KEYBOARD_GOX): $(KEYBOARD_OBJ) | $(BUILD_DIR)
	mkdir -p $(dir $(KEYBOARD_GOX))
	$(OBJCOPY) -j .go_export $(KEYBOARD_OBJ) $(KEYBOARD_GOX)

$(MEM_OBJ): $(MEM_SRCS) | $(BUILD_DIR)
	$(GCCGO) $(GCCGOFLAGS) -static -Werror -nostdlib -nostartfiles -nodefaultlibs \
		-fgo-pkgpath=$(MEM_IMPORT) \
		-c $(MEM_SRCS) -o $(MEM_OBJ)

$(MEM_GOX): $(MEM_OBJ) | $(BUILD_DIR)
	mkdir -p $(dir $(MEM_GOX))
	$(OBJCOPY) -j .go_export $(MEM_OBJ) $(MEM_GOX)

$(FS_OBJ): $(FS_SRCS) $(MEM_GOX) | $(BUILD_DIR)
	$(GCCGO) $(GCCGOFLAGS) -static -Werror -nostdlib -nostartfiles -nodefaultlibs \
		-I $(BUILD_DIR) \
		-fgo-pkgpath=$(FS_IMPORT) \
		-c $(FS_SRCS) -o $(FS_OBJ)

$(FS_GOX): $(FS_OBJ) | $(BUILD_DIR)
	mkdir -p $(dir $(FS_GOX))
	$(OBJCOPY) -j .go_export $(FS_OBJ) $(FS_GOX)

# --- 6. Compile shell.go (package shell) with gccgo ---
$(SHELL_OBJ): $(SHELL_SRCS) $(TERMINAL_GOX) $(MEM_GOX) $(FS_GOX) | $(BUILD_DIR)
	$(GCCGO) $(GCCGOFLAGS) -static -Werror -nostdlib -nostartfiles -nodefaultlibs \
		-I $(BUILD_DIR) \
		-fgo-pkgpath=$(SHELL_IMPORT) \
		-c $(SHELL_SRCS) -o $(SHELL_OBJ)

# --- 7. Extract .go_export into shell.gox ---
$(SHELL_GOX): $(SHELL_OBJ) | $(BUILD_DIR)
	mkdir -p $(dir $(SHELL_GOX))
	$(OBJCOPY) -j .go_export $(SHELL_OBJ) $(SHELL_GOX)

# --- Scheduler ---
$(SCHEDULER_OBJ): $(SCHEDULER_SRCS) | $(BUILD_DIR)
	$(GCCGO) $(GCCGOFLAGS) -static -Werror -nostdlib -nostartfiles -nodefaultlibs \
		-fgo-pkgpath=$(SCHEDULER_IMPORT) \
		-c $(SCHEDULER_SRCS) -o $(SCHEDULER_OBJ)

$(SCHEDULER_GOX): $(SCHEDULER_OBJ) | $(BUILD_DIR)
	mkdir -p $(dir $(SCHEDULER_GOX))
	$(OBJCOPY) -j .go_export $(SCHEDULER_OBJ) $(SCHEDULER_GOX)

$(SCH_SWITCH_OBJ): $(SCH_SWITCH_SRC) | $(BUILD_DIR)
	$(AS) $(SCH_SWITCH_SRC) -o $(SCH_SWITCH_OBJ)

# --- 8. Compile kernel.go (package kernel, imports "github.com/dmarro89/go-dav-os/terminal") ---
$(KERNEL_OBJ): $(KERNEL_SRCS) $(TERMINAL_GOX) $(KEYBOARD_GOX) $(SHELL_GOX) $(MEM_GOX) $(FS_GOX) $(SCHEDULER_GOX) | $(BUILD_DIR)
	$(GCCGO) $(GCCGOFLAGS) -static -Werror -nostdlib -nostartfiles -nodefaultlibs \
		-I $(BUILD_DIR) \
		-c $(KERNEL_SRCS) -o $(KERNEL_OBJ)

# -----------------------
# Link: boot.o + kernel.o -> kernel.elf
# -----------------------
$(KERNEL_ELF): $(BOOT_OBJ) $(TERMINAL_OBJ) $(KEYBOARD_OBJ) $(SHELL_OBJ) $(MEM_OBJ) $(FS_OBJ) $(SCHEDULER_OBJ) $(SCH_SWITCH_OBJ) $(KERNEL_OBJ) $(LINKER_SCRIPT)
	$(GCC) -T $(LINKER_SCRIPT) -o $(KERNEL_ELF) \
		-ffreestanding -O2 -nostdlib \
		$(BOOT_OBJ) $(TERMINAL_OBJ) $(KEYBOARD_OBJ) $(SHELL_OBJ) $(MEM_OBJ) $(FS_OBJ) $(SCHEDULER_OBJ) $(SCH_SWITCH_OBJ) $(KERNEL_OBJ) -lgcc

# -----------------------
# ISO with GRUB
# -----------------------
$(ISO_DIR)/boot/grub:
	mkdir -p $(ISO_DIR)/boot/grub

$(ISO_DIR)/boot/kernel.elf: $(KERNEL_ELF) $(GRUB_CFG) | $(ISO_DIR)/boot/grub
	cp $(KERNEL_ELF) $(ISO_DIR)/boot/kernel.elf
	cp $(GRUB_CFG) $(ISO_DIR)/boot/grub/grub.cfg

$(ISO_IMAGE): $(ISO_DIR)/boot/kernel.elf
	$(GRUBMKRESCUE) -o $(ISO_IMAGE) $(ISO_DIR)

# -----------------------
# Docker helpers
# -----------------------
docker-image:
	docker build --platform=$(DOCKER_PLATFORM) -t $(DOCKER_IMAGE) .

docker-run: docker-image
	docker run ${DOCKER_RUN_FLAGS} --rm --platform=$(DOCKER_PLATFORM) \
	  -v "$(CURDIR)":/work -w /work $(DOCKER_IMAGE) \
	  make run

docker-build-only: docker-image
	docker run --rm --platform=$(DOCKER_PLATFORM) \
	  -v "$(CURDIR)":/work -w /work $(DOCKER_IMAGE) \
	  make

docker-shell: docker-image
	docker run -it --rm --platform=$(DOCKER_PLATFORM) \
	  -v "$(CURDIR)":/work -w /work $(DOCKER_IMAGE) bash
