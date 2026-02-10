# Recommended Learning Path

If this is your first kernel project, use this order:

1. `00-introduction/README.md`
2. `00-introduction/glossary.md`
3. `01-overview/architecture.md`
4. `02-boot/boot-and-grub.md`
5. `03-kernel-core/main-loop.md`
6. `03-kernel-core/theory-reference.md`
7. `03-kernel-core/interrupts-and-syscalls.md`
8. `03-kernel-core/scheduler-and-tasks.md`
9. `04-memory/page-frame-allocator.md`
10. `05-io/vga-terminal.md` and `05-io/ps2-keyboard.md`
11. `06-filesystem/filesystem-in-memory.md` and `06-filesystem/fat16.md`
12. `07-shell/shell-commands.md`
13. `08-build-and-test/build-run.md`

## Practical method

For each page:

1. Read the "What it does" section
2. Run shell commands from examples (when available)
3. Open the referenced source files
4. Try a small change and rebuild

This sequence reduces the risk of getting lost in low-level details too early.
