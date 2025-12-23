CROSS    ?= i686-elf

AS       := $(CROSS)-as
GCC      := $(CROSS)-gcc
GCCGO    := $(CROSS)-gccgo
OBJCOPY  := $(CROSS)-objcopy
GRUB_CFG      := iso/grub/grub.cfg

GRUBMKRESCUE  := grub-mkrescue
QEMU          := qemu-system-i386

DOCKER_PLATFORM := linux/amd64
DOCKER_IMAGE    := go-dav-os-toolchain
DOCKER_RUN_FLAGS=-it

BUILD_DIR := build
ISO_DIR   := $(BUILD_DIR)/isodir

KERNEL_ELF := $(BUILD_DIR)/kernel.elf
ISO_IMAGE   := $(BUILD_DIR)/dav-go-os.iso

BOOT_SRC      := boot/boot.s
LINKER_SCRIPT := boot/linker.ld

MODPATH          := github.com/dmarro89/go-dav-os
TERMINAL_IMPORT  := $(MODPATH)/terminal
KEYBOARD_IMPORT  := $(MODPATH)/keyboard

KERNEL_SRCS := $(wildcard kernel/*.go)
KERNEL_SRC   := kernel/kernel.go kernel/idt.go
TERMINAL_SRC := terminal/terminal.go
KEYBOARD_SRC := keyboard/keyboard.go keyboard/layout.go
KEYBOARD_SRCS := $(wildcard keyboard/*.go)

BOOT_OBJ   := $(BUILD_DIR)/boot.o
KERNEL_OBJ := $(BUILD_DIR)/kernel.o
TERMINAL_OBJ := $(BUILD_DIR)/terminal.o
TERMINAL_GOX := $(BUILD_DIR)/github.com/dmarro89/go-dav-os/terminal.gox
KEYBOARD_OBJ   := $(BUILD_DIR)/keyboard.o
KEYBOARD_GOX   := $(BUILD_DIR)/github.com/dmarro89/go-dav-os/keyboard.gox

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
$(BOOT_OBJ): $(BOOT_SRC) | $(BUILD_DIR)
	$(AS) $(BOOT_SRC) -o $(BOOT_OBJ)

# --- 2. Compile terminal.go (package terminal) with gccgo ---
$(TERMINAL_OBJ): $(TERMINAL_SRC) | $(BUILD_DIR)
	$(GCCGO) -static -Werror -nostdlib -nostartfiles -nodefaultlibs \
		-fgo-pkgpath=$(TERMINAL_IMPORT) \
		-c $(TERMINAL_SRC) -o $(TERMINAL_OBJ)

# --- 3. Extract .go_export into terminal.gox ---
$(TERMINAL_GOX): $(TERMINAL_OBJ) | $(BUILD_DIR)
	mkdir -p $(dir $(TERMINAL_GOX))
	$(OBJCOPY) -j .go_export $(TERMINAL_OBJ) $(TERMINAL_GOX)

# --- 4. Compile keyboard.go and layout.go (package keyboard) with gccgo ---
$(KEYBOARD_OBJ): $(KEYBOARD_SRCS) | $(BUILD_DIR)
	$(GCCGO) -static -Werror -nostdlib -nostartfiles -nodefaultlibs \
		-fgo-pkgpath=$(KEYBOARD_IMPORT) \
		-c $(KEYBOARD_SRC) -o $(KEYBOARD_OBJ)

# --- 5. Extract .go_export into keyboard.gox ---
$(KEYBOARD_GOX): $(KEYBOARD_OBJ) | $(BUILD_DIR)
	mkdir -p $(dir $(KEYBOARD_GOX))
	$(OBJCOPY) -j .go_export $(KEYBOARD_OBJ) $(KEYBOARD_GOX)

# --- 7. Compile kernel.go (package kernel, imports "github.com/dmarro89/go-dav-os/terminal") ---
$(KERNEL_OBJ): $(KERNEL_SRCS) $(TERMINAL_GOX) $(KEYBOARD_GOX) | $(BUILD_DIR)
	$(GCCGO) -static -Werror -nostdlib -nostartfiles -nodefaultlibs \
		-I $(BUILD_DIR) \
		-c $(KERNEL_SRCS) -o $(KERNEL_OBJ)

# -----------------------
# Link: boot.o + kernel.o -> kernel.elf
# -----------------------
$(KERNEL_ELF): $(BOOT_OBJ) $(TERMINAL_OBJ) $(KEYBOARD_OBJ) $(KERNEL_OBJ) $(LINKER_SCRIPT)
	$(GCC) -T $(LINKER_SCRIPT) -o $(KERNEL_ELF) \
		-ffreestanding -O2 -nostdlib \
		$(BOOT_OBJ) $(TERMINAL_OBJ) $(KEYBOARD_OBJ) $(KERNEL_OBJ) -lgcc

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
